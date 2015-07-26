package main

import (
	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"gopkg.in/olivere/elastic.v2"
)

func main() {
	client, err := elastic.NewClient()
	if err != nil {
		panic("Unable to connect to elasticsearch")
	}

	exists, err := client.IndexExists("files").Do()
	if err != nil {
		panic("Failed checking index existence")
	}

	if !exists {
		createStatus, err := client.CreateIndex("files").Do()
		if err != nil {
			panic("Failed creating index")
		}
		if !createStatus.Acknowledged {
			fmt.Println("Index create not acknowledged")
		}
	} else {
		fmt.Println("Index exists")
	}

	session := db.RSessionMust()
	res, err := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2datafile"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("datafile_id", r.Table("datafiles")).
		Zip().
		Limit(10).
		Run(session)
	if err != nil {
		panic(fmt.Sprintf("Unable to query database: %s", err))
	}
	defer res.Close()

	var df schema.File
	count := 0
	maxCount := 100
	bulkReq := client.Bulk()
	for res.Next(&df) {
		if count < maxCount {
			indexReq := elastic.NewBulkIndexRequest().Index("files").Type("file").Id(df.ID).Doc(df)
			bulkReq = bulkReq.Add(indexReq)
			count++
		} else {
			count = 0
			_, err := bulkReq.Do()
			if err != nil {
				panic(fmt.Sprintf("bulkreq failed: %s", err))
			}
			bulkReq = client.Bulk()
		}
	}

	if count != 0 {
		bulkReq.Do()
	}
}
