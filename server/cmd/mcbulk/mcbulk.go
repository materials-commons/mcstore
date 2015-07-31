package main

import (
	"errors"
	"fmt"

	"io/ioutil"

	"bufio"
	"os"

	"strings"

	"os/exec"

	"github.com/codegangsta/cli"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
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
	     		"process_id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	             },
	             "project_id": {
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

func main() {
	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Authors = []cli.Author{
		{
			Name:  "V. Glenn Tarcea",
			Email: "gtarcea@umich.edu",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "es-url",
			Value:  "http://localhost:9500",
			Usage:  "Elasticsearch server URL",
			EnvVar: "MC_ES_URL",
		},

		cli.StringFlag{
			Name:   "db-connection",
			Value:  "localhost:30815",
			Usage:  "RethinkDB connection string",
			EnvVar: "MCDB_CONNECTION",
		},

		cli.StringFlag{
			Name:   "db-name",
			Value:  "materialscommons",
			Usage:  "Database to index",
			EnvVar: "MCDB_NAME",
		},

		cli.StringFlag{
			Name:   "mc-dir",
			Value:  "/mcfs/data/test",
			Usage:  "Path to data directory",
			EnvVar: "MCDIR",
		},

		cli.BoolFlag{
			Name:  "create-index",
			Usage: "Whether the index should be recreated",
		},

		cli.BoolTFlag{
			Name:  "processes",
			Usage: "Index processes",
		},

		cli.BoolTFlag{
			Name:  "files",
			Usage: "Index files",
		},

		cli.BoolTFlag{
			Name:  "samples",
			Usage: "Index samples",
		},

		cli.BoolTFlag{
			Name:  "projects",
			Usage: "Index projects",
		},

		cli.BoolTFlag{
			Name:  "users",
			Usage: "Index users",
		},
	}

	app.Action = mcbulkCLI
	app.Run(os.Args)
}

func mcbulkCLI(c *cli.Context) {
	setupConfig(c)
	runCommands(c)
}

func setupConfig(c *cli.Context) {
	esurl := c.String("es-url")
	config.Set("MC_ES_URL", esurl)

	dbname := c.String("db-name")
	config.Set("MCDB_NAME", dbname)

	dbcon := c.String("db-connection")
	config.Set("MCDB_CONNECTION", dbcon)
}

func runCommands(c *cli.Context) {
	esurl := esURL()
	fmt.Println("Elasticsearch URL:", esurl)
	client, err := elastic.NewClient(elastic.SetURL(esurl))
	if err != nil {
		panic("Unable to connect to elasticsearch")
	}

	session := db.RSessionMust()

	if c.Bool("create-index") {
		createIndex(client)
	}

	if c.BoolT("files") {
		loadFiles(client, session)
	}

	if c.BoolT("users") {
		loadUsers(client, session)
	}

	if c.BoolT("projects") {
		loadProjects(client, session)
	}

	if c.BoolT("samples") {
		loadSamples(client, session)
	}

	if c.BoolT("processes") {
		loadProcesses(client, session)
	}
}

func esURL() string {
	if esURL := config.GetString("MC_ES_URL"); esURL != "" {
		return esURL
	}
	return "http://localhost:9200"
}

func createIndex(client *elastic.Client) {
	fmt.Println("Creating index mc...")
	exists, err := client.IndexExists("mc").Do()
	if err != nil {
		panic("  Failed checking index existence")
	}

	if exists {
		fmt.Println("  Index exists deleting old one")
		client.DeleteIndex("mc").Do()
	}

	createStatus, err := client.CreateIndex("mc").Body(mappings).Do()
	if err != nil {
		fmt.Println("  Failed creating index: ", err)
		os.Exit(1)
	}

	if !createStatus.Acknowledged {
		fmt.Println("  Index create not acknowledged")
	}

	fmt.Println("Done.")
}

func loadFiles(client *elastic.Client, session *r.Session) {
	var df doc.File
	renameDirPath := func(row r.Term) interface{} {
		return row.Merge(map[string]interface{}{
			"right": map[string]interface{}{
				"path": row.Field("right").Field("name"),
			},
		})
	}

	tagsAndNotes := func(row r.Term) interface{} {
		return map[string]interface{}{
			"tags": r.Table("tag2item").GetAllByIndex("item_id", row.Field("id")).
				Pluck("tag_id").CoerceTo("ARRAY"),
			"notes": r.Table("note2item").GetAllByIndex("item_id", row.Field("id")).
				EqJoin("note_id", r.Table("notes")).Zip().CoerceTo("ARRAY"),
		}
	}

	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2datafile"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("datafile_id", r.Table("datadir2datafile"), r.EqJoinOpts{Index: "datafile_id"}).Zip().
		EqJoin("datadir_id", r.Table("datadirs")).
		Map(renameDirPath).
		Zip().
		EqJoin("datafile_id", r.Table("datafiles")).Zip().
		//Filter(r.Row.Field("id").Eq("184e5b21-b86a-4fd0-97ea-98c726a9787b")).
		//Filter(r.Row.Field("id").Eq("b20cde2d-350b-4bc4-8700-e42352bb70df")).
		Merge(tagsAndNotes)

	filesIndexer := &search.Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			dfile := item.(*doc.File)
			return dfile.ID
		},
		Apply: func(item interface{}) {
			dfile := item.(*doc.File)
			dfile.Contents = readContents(dfile.ID, dfile.MediaType.Mime, dfile.Name, dfile.Size)
		},
		Client:   client,
		Session:  session,
		MaxCount: 10,
	}
	fmt.Println("Indexing files...")
	if err := filesIndexer.Do("files", df); err != nil {
		fmt.Println("  Indexing files failed:", err)
		fmt.Println("  Some files may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadUsers(client *elastic.Client, session *r.Session) {
	var u schema.User
	rql := r.Table("users")

	usersIndexer := &search.Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			user := item.(*schema.User)
			return user.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}

	fmt.Println("Indexing users...")
	if err := usersIndexer.Do("users", u); err != nil {
		fmt.Println("  Indexing users failed:", err)
		fmt.Println("  Some users may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadProjects(client *elastic.Client, session *r.Session) {
	var p schema.Project
	rql := r.Table("projects")
	projectsIndexer := &search.Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			project := item.(*schema.Project)
			return project.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}

	fmt.Println("Indexing projects...")
	if err := projectsIndexer.Do("projects", p); err != nil {
		fmt.Println("  Indexing projects failed:", err)
		fmt.Println("  Some projects may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadSamples(client *elastic.Client, session *r.Session) {
	var sample doc.Sample
	propertiesAndFiles := func(row r.Term) interface{} {
		return map[string]interface{}{
			"properties": r.Table("sample2propertyset").
				GetAllByIndex("sample_id", row.Field("sample_id")).
				EqJoin("property_set_id", r.Table("propertyset2property"), r.EqJoinOpts{Index: "property_set_id"}).
				Zip().
				EqJoin("property_id", r.Table("properties")).Zip().Pluck("attribute", "name").
				CoerceTo("ARRAY"),
			"files": r.Table("sample2datafile").GetAllByIndex("sample_id", row.Field("sample_id")).
				EqJoin("datafile_id", r.Table("datafiles")).Zip().CoerceTo("ARRAY"),
		}
	}
	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2sample"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("sample_id", r.Table("samples")).Zip().
		Merge(propertiesAndFiles)

	samplesIndexer := &search.Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			s := item.(*doc.Sample)
			return s.SampleID
		},
		Apply: func(item interface{}) {
			s := item.(*doc.Sample)
			for i, _ := range s.Files {
				f := s.Files[i]
				s.Files[i].Contents = readContents(f.DataFileID, f.MediaType.Mime, f.Name, f.Size)
			}
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}

	fmt.Println("Indexing samples...")
	if err := samplesIndexer.Do("samples", sample); err != nil {
		fmt.Println("  Indexing samples failed:", err)
		fmt.Println("  Some samples may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadProcesses(client *elastic.Client, session *r.Session) {
	var process doc.Process

	getSetup := func(row r.Term) interface{} {
		return map[string]interface{}{
			"setup": r.Table("process2setup").GetAllByIndex("process_id", row.Field("process_id")).
				EqJoin("setup_id", r.Table("setupproperties"), r.EqJoinOpts{Index: "setup_id"}).
				Zip().CoerceTo("ARRAY"),
		}
	}

	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2process"), r.EqJoinOpts{Index: "project_id"}).
		Zip().
		EqJoin("process_id", r.Table("processes")).Zip().
		Merge(getSetup)

	processesIndexer := &search.Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			s := item.(*doc.Process)
			return s.ProcessID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}

	fmt.Println("Indexing processes...")
	if err := processesIndexer.Do("processes", process); err != nil {
		fmt.Println("  Indexing processes failed:", err)
		fmt.Println("  Some processes may not have been indexed.")
	}
	fmt.Println("Done.")
}

const twoMeg = 2 * 1024 * 1024

func readContents(fileID, mimeType, name string, size int64) string {
	switch mimeType {
	case "text/csv":
		//fmt.Println("Reading csv file: ", fileID, name, size)
		if contents, err := readCSVLines(fileID); err == nil {
			return contents
		}
	case "text/plain":
		if size > twoMeg {
			return ""
		}
		//fmt.Println("Reading text file: ", fileID, name, size)
		if contents, err := ioutil.ReadFile(app.MCDir.FilePath(fileID)); err == nil {
			return string(contents)
		}
	default:
		if _, ok := tikableMediaTypes[mimeType]; ok {
			if contents := extractUsingTika(fileID, mimeType, name, size); contents != "" {
				return contents
			}
		}
	}
	return ""
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
		//fmt.Println("readCSVLines no data")
		return "", errors.New("No data")
	} else {
		//fmt.Println("readCSVLines failed to open", err)
		return "", err
	}
}

func extractUsingTika(fileID, mimeType, name string, size int64) string {
	if size > twoMeg {
		return ""
	}

	out, err := exec.Command("tika.sh", "--text", app.MCDir.FilePath(fileID)).Output()
	if err != nil {
		fmt.Println("Tika failed for:", fileID, name, mimeType)
		fmt.Println("exec failed:", err)
		return ""
	}

	return string(out)
}
