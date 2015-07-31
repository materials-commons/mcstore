package search

import (
	"reflect"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"gopkg.in/olivere/elastic.v2"
)

type Indexer struct {
	RQL      r.Term
	GetID    func(item interface{}) string
	Apply    func(item interface{})
	Client   *elastic.Client
	Session  *r.Session
	MaxCount int
}

func (i *Indexer) Do(itype string, what interface{}) error {
	res, err := i.RQL.Run(i.Session)
	if err != nil {
		app.Log.Errorf("Failed to run query: %s", err)
		return err
	}
	defer res.Close()

	total := 0
	count := 0
	bulkReq := i.Client.Bulk()
	elementType := reflect.TypeOf(what)
	result := reflect.New(elementType)
	for res.Next(result.Interface()) {
		if i.Apply != nil {
			i.Apply(result.Interface())
		}

		if count < i.MaxCount {
			id := i.GetID(result.Interface())
			indexReq := elastic.NewBulkIndexRequest().Index("mc").Type(itype).Id(id).Doc(result.Interface())
			bulkReq = bulkReq.Add(indexReq)
			count++
			total++
		} else {
			count = 0
			resp, err := bulkReq.Do()
			if err != nil {
				app.Log.Errorf("bulkreq failed: %s %#v\n", err, resp)
				return err
			}
		}
		result = reflect.New(elementType)
	}

	if res.Err() != nil {
		app.Log.Errorf("RethinkDB cursor error", err)
	}

	if count != 0 {
		app.Log.Infof("  Indexed %d %s...\n", total, itype)
		bulkReq.Do()
	}

	return nil
}
