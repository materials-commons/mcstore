package search

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
	"gopkg.in/olivere/elastic.v2"
)

func getProcessesSetup(row r.Term) interface{} {
	return map[string]interface{}{
		"setup": r.Table("process2setup").GetAllByIndex("process_id", row.Field("process_id")).
			EqJoin("setup_id", r.Table("setupproperties"), r.EqJoinOpts{Index: "setup_id"}).
			Zip().CoerceTo("ARRAY"),
	}
}

func NewProcessesIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2process"), r.EqJoinOpts{Index: "project_id"}).
		Zip().
		EqJoin("process_id", r.Table("processes")).Zip().
		Merge(getProcessesSetup)

	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			s := item.(*doc.Process)
			return s.ProcessID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}

func NewSingleProcessIndexer(client *elastic.Client, session *r.Session, processID string) *Indexer {
	rql := r.Table("project2process").GetAllByIndex("process_id", processID).
		Zip().
		EqJoin("process_id", r.Table("processes")).Zip().
		Merge(getProcessesSetup)

	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			s := item.(*doc.Process)
			return s.ProcessID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}
