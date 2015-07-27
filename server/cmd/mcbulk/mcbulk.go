package main

import (
	"errors"
	"fmt"

	"io/ioutil"

	"bufio"
	"os"

	"strings"

	r "github.com/dancannon/gorethink"
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
	              },
	              "name": {
	               	  "type": "string",
	                  "index": "not_analyzed"
	              }
	         }
	     }
	}
}
`

func main() {
	client, err := elastic.NewClient()
	if err != nil {
		panic("Unable to connect to elasticsearch")
	}

	exists, err := client.IndexExists("mc").Do()
	if err != nil {
		panic("Failed checking index existence")
	}

	if exists {
		client.DeleteIndex("mc").Do()
	}

	createStatus, err := client.CreateIndex("mc").Body(mappings).Do()
	if err != nil {
		panic("Failed creating index")
	}
	if !createStatus.Acknowledged {
		fmt.Println("Index create not acknowledged")
	}

	session := db.RSessionMust()
	res, err := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2datafile"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("datafile_id", r.Table("datadir2datafile"), r.EqJoinOpts{Index: "datafile_id"}).Zip().
		EqJoin("datafile_id", r.Table("datafiles")).Zip().
		Run(session)
	if err != nil {
		panic(fmt.Sprintf("Unable to query database: %s", err))
	}
	defer res.Close()

	var df schema.File
	count := 0
	maxCount := 10
	bulkReq := client.Bulk()
	for res.Next(&df) {
		readContents(&df)
		if count < maxCount {
			indexReq := elastic.NewBulkIndexRequest().Index("mc").Type("files").Id(df.ID).Doc(df)
			bulkReq = bulkReq.Add(indexReq)
			count++
		} else {
			count = 0
			resp, err := bulkReq.Do()
			if err != nil {
				fmt.Printf("bulkreq failed: %s\n", err)
				fmt.Printf("%#v\n", resp)
				break
			}
			//bulkReq = client.Bulk()
		}
	}

	if count != 0 {
		bulkReq.Do()
	}
}

const twoMeg = 2 * 1024 * 1024

func readContents(file *schema.File) {
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
