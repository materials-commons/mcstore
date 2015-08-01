package search

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"gopkg.in/olivere/elastic.v2"
)

func NewProjectsIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("projects")
	indexer := defaultProjectIndexer(client, session)
	indexer.RQL = rql
	return indexer
}

func NewSingleProjectIndexer(client *elastic.Client, session *r.Session, projectID string) *Indexer {
	rql := r.Table("projects").GetAll(projectID)
	indexer := defaultProjectIndexer(client, session)
	indexer.RQL = rql
	return indexer
}

func defaultProjectIndexer(client *elastic.Client, session *r.Session) *Indexer {
	return &Indexer{
		GetID: func(item interface{}) string {
			project := item.(*schema.Project)
			return project.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}
