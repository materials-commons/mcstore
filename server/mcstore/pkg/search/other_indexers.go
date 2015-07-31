package search

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"gopkg.in/olivere/elastic.v2"
)

func NewProjectsIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("projects")
	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			project := item.(*schema.Project)
			return project.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}

func NewUsersIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("users")

	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			user := item.(*schema.User)
			return user.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}
