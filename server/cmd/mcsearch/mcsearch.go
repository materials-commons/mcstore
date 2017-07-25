package main

import (
	"fmt"
	"gopkg.in/olivere/elastic.v5"
	"os"
)

func main() {
	client, err := elastic.NewClient()
	if err != nil {
		panic("Unable to connect to elasticsearch")
	}

	termQueryProj := elastic.NewTermQuery("project_id", "d232df78-cbe2-4561-a958-7fd45b87601d")
	termQueryName := elastic.NewTermQuery("name", "2-30k.tif")
	boolQuery := elastic.NewBoolQuery()
	boolQuery = boolQuery.Must(termQueryProj, termQueryName)
	results, err := client.Search().
		Index("mc").Type("files").
		Query(&boolQuery).
		Do()
	if err != nil {
		fmt.Println("Search failed: ", err)
		os.Exit(1)
	}

	fmt.Println("Found: ", results.TotalHits())
}
