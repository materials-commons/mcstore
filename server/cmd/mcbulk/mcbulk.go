package main

import (
	"errors"
	"fmt"

	"io/ioutil"

	"bufio"
	"os"

	"strings"

	"os/exec"

	"time"

	"reflect"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"gopkg.in/olivere/elastic.v2"
)

var mappings string = `
{
	"mappings": {
	     "files": {
	         "properties": {
	              "project_id": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              },
	              "project": {
	              		"type": "string",
	              		"index": "not_analyzed"
	              },
	              "datadir_id": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              },
	              "id": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              },
	              "usesid": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              }
	         }
	     },
	     "projects": {
	     	"properties": {
	     		"id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	             },
	             "datadir": {
	             	"type": "string",
	             	"index": "not_analyzed"
	             }
	     	}
	     },
	     "samples": {
	     	"properties":{
	     	    "id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	            },
	        	"project_id": {
	                "type": "string",
	                "index": "not_analyzed"
	        	},
	        	"sample_id": {
	        		"type": "string",
	        		"index": "not_analyzed"
	        	}
	        }
	     },
	     "processes": {
	     	"properties":{
	     		"id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	             }
	        }
	     },
		"users": {
			"properties":{
	     		"id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	             }
	        }
	    }
	}
}
`

var tikableMediaTypes map[string]bool = map[string]bool{
	"application/msword":                                                        true,
	"application/pdf":                                                           true,
	"application/rtf":                                                           true,
	"application/vnd.ms-excel":                                                  true,
	"application/vnd.ms-office":                                                 true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.ms-powerpoint.presentation.macroEnabled.12":                true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.sealedmedia.softseal.pdf":                                  true,
	"text/plain; charset=utf-8":                                                 true,
}

//var onlyHeader map[string]bool = map[string]bool{
//	"application/vnd.ms-excel":                                          true,
//	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
//}

func main() {
	esurl := esURL()
	fmt.Println("Elasticsearch URL:", esurl)
	client, err := elastic.NewClient(elastic.SetURL(esurl))
	if err != nil {
		panic("Unable to connect to elasticsearch")
	}

	session := db.RSessionMust()

	createIndex(client)
	loadFiles(client, session)
	loadUsers(client, session)
	loadProjects(client, session)
	loadSamples(client, session)
}

func esURL() string {
	if esURL := config.GetString("MC_ES_URL"); esURL != "" {
		return esURL
	}
	return "http://localhost:9200"
}

func createIndex(client *elastic.Client) {
	exists, err := client.IndexExists("mc").Do()
	if err != nil {
		panic("Failed checking index existence")
	}

	if exists {
		client.DeleteIndex("mc").Do()
	}

	createStatus, err := client.CreateIndex("mc").Body(mappings).Do()
	if err != nil {
		fmt.Println("Failed creating index: ", err)
		os.Exit(1)
	}
	if !createStatus.Acknowledged {
		fmt.Println("Index create not acknowledged")
	}
}

type TagID struct {
	TagID string `gorethink:"tag_id" json:"tag"`
}

type File struct {
	schema.File
	Tags      []TagID `gorethink:"tags" json:"tags"`
	DataDirID string  `gorethink:"datadir_id" json:"datadir_id"` // Directory file is located in
	ProjectID string  `gorethink:"project_id" json:"project_id"` // Project file is in
	Contents  string  `gorethink:"-" json:"contents"`            // Contents of the file (text only)
}

func loadFiles(client *elastic.Client, session *r.Session) {
	var df File
	renameDirPath := func(row r.Term) interface{} {
		return row.Merge(map[string]interface{}{
			"right": map[string]interface{}{
				"path": row.Field("right").Field("name"),
			},
		})
	}

	fileTags := func(row r.Term) interface{} {
		return map[string]interface{}{
			"tags": r.Table("tag2item").GetAllByIndex("item_id", row.Field("id")).
				Pluck("tag_id").CoerceTo("ARRAY"),
		}
	}

	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2datafile"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("datafile_id", r.Table("datadir2datafile"), r.EqJoinOpts{Index: "datafile_id"}).Zip().
		EqJoin("datadir_id", r.Table("datadirs")).
		Map(renameDirPath).
		Zip().
		EqJoin("datafile_id", r.Table("datafiles")).Zip().
		Merge(fileTags)

	fileIndexer := &indexer{
		rql: rql,
		getID: func(item interface{}) string {
			dfile := item.(*File)
			return dfile.ID
		},
		apply: func(item interface{}) {
			dfile := item.(*File)
			readContents(dfile)
		},
		client:   client,
		session:  session,
		maxCount: 10,
	}
	fmt.Println("Indexing files...")
	fileIndexer.Do("files", df)
	fmt.Println("Done.")
}

func loadUsers(client *elastic.Client, session *r.Session) {
	var u schema.User
	rql := r.Table("users")

	userIndexer := &indexer{
		rql: rql,
		getID: func(item interface{}) string {
			user := item.(*schema.User)
			return user.ID
		},
		client:   client,
		session:  session,
		maxCount: 1000,
	}

	fmt.Println("Indexing users...")
	userIndexer.Do("users", u)
	fmt.Println("Done.")
}

func loadProjects(client *elastic.Client, session *r.Session) {
	var p schema.Project
	rql := r.Table("projects")
	projectIndexer := &indexer{
		rql: rql,
		getID: func(item interface{}) string {
			project := item.(*schema.Project)
			return project.ID
		},
		client:   client,
		session:  session,
		maxCount: 1000,
	}

	fmt.Println("Indexing projects...")
	projectIndexer.Do("projects", p)
	fmt.Println("Done.")
}

type Property struct {
	Attribute string `gorethink:"attribute" json:"attribute"`
	Name      string `gorethink:"name" json:"name"`
}

type Sample struct {
	ID          string     `gorethink:"id" json:"id"`
	Description string     `gorethink:"description" json:"description"`
	Birthtime   time.Time  `gorethink:"birthtime" json:"birthtime"`
	MTime       time.Time  `gorethink:"mtime" json:"mtime"`
	Owner       string     `gorethink:"owner" json:"owner"`
	Name        string     `gorethink:"name" json:"name"`
	ProjectID   string     `gorethink:"project_id" json:"project_id"`
	SampleID    string     `gorethink:"sample_id" json:"sample_id"`
	Properties  []Property `gorethink:"properties" json:"properties"`
}

func loadSamples(client *elastic.Client, session *r.Session) {
	var sample Sample
	getProperties := func(row r.Term) interface{} {
		return map[string]interface{}{
			"properties": r.Table("sample2propertyset").
				GetAllByIndex("sample_id", row.Field("sample_id")).
				EqJoin("property_set_id", r.Table("propertyset2property"), r.EqJoinOpts{Index: "property_set_id"}).
				Zip().
				EqJoin("property_id", r.Table("properties")).Zip().Pluck("attribute", "name").
				CoerceTo("ARRAY"),
		}
	}
	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2sample"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("sample_id", r.Table("samples")).Zip().
		Merge(getProperties)

	sampleIndexer := &indexer{
		rql: rql,
		getID: func(item interface{}) string {
			s := item.(*Sample)
			return s.SampleID
		},
		client:   client,
		session:  session,
		maxCount: 1000,
	}

	fmt.Println("Indexing samples...")
	sampleIndexer.Do("samples", sample)
	fmt.Println("Done.")
}

type indexer struct {
	rql      r.Term
	getID    func(item interface{}) string
	apply    func(item interface{})
	client   *elastic.Client
	session  *r.Session
	maxCount int
}

func (i *indexer) Do(itype string, what interface{}) {
	res, err := i.rql.Run(i.session)
	if err != nil {
		fmt.Println("Failed to run query:", err)
		os.Exit(1)
	}
	defer res.Close()

	total := 0
	count := 0
	bulkReq := i.client.Bulk()
	elementType := reflect.TypeOf(what)
	result := reflect.New(elementType)
	for res.Next(result.Interface()) {
		if i.apply != nil {
			i.apply(result.Interface())
		}

		if count < i.maxCount {
			id := i.getID(result.Interface())
			indexReq := elastic.NewBulkIndexRequest().Index("mc").Type(itype).Id(id).Doc(result.Interface())
			bulkReq = bulkReq.Add(indexReq)
			count++
			total++
		} else {
			fmt.Printf("  Indexed %d...\n", total)
			count = 0
			resp, err := bulkReq.Do()
			if err != nil {
				fmt.Printf("bulkreq failed: %s\n", err)
				fmt.Printf("%#v\n", resp)
				return
			}
		}
		result = reflect.New(elementType)
	}

	if res.Err() != nil {
		fmt.Println("res err", err)
	}

	if count != 0 {
		fmt.Printf("  Indexed %d %s...\n", total, itype)
		bulkReq.Do()
	}
}

const twoMeg = 2 * 1024 * 1024

func readContents(file *File) {
	switch file.MediaType.Mime {
	case "text/csv":
		//fmt.Println("Reading csv file: ", file.ID, file.Name, file.Size)
		if contents, err := readCSVLines(file.ID); err == nil {
			file.Contents = string(contents)
		}
	case "text/plain":
		if file.Size > twoMeg {
			return
		}
		//fmt.Println("Reading text file: ", file.ID, file.Name, file.Size)
		if contents, err := ioutil.ReadFile(app.MCDir.FilePath(file.ID)); err == nil {
			file.Contents = string(contents)
		}
	default:
		if _, ok := tikableMediaTypes[file.MediaType.Mime]; ok {
			if contents := extractUsingTika(file); contents != "" {
				file.Contents = contents
			}
		}
	}
}

func readCSVLines(fileID string) (string, error) {
	if file, err := os.Open(app.MCDir.FilePath(fileID)); err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			text := scanner.Text()
			if text != "" && !strings.HasPrefix(text, "#") {
				return text, nil
			}
		}
		return "", errors.New("No data")
	} else {
		return "", err
	}
}

func extractUsingTika(file *File) string {
	if file.Size > twoMeg {
		return ""
	}

	out, err := exec.Command("tika.sh", "--text", app.MCDir.FilePath(file.ID)).Output()
	if err != nil {
		fmt.Println("Tika failed for:", file.Name, file.ID, file.MediaType.Mime)
		fmt.Println("exec failed:", err)
		return ""
	}

	return string(out)
}
