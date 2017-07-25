package search

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"gopkg.in/olivere/elastic.v5"
)

func NewUsersIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("users")
	indexer := defaultUsersIndexer(client, session)
	indexer.RQL = rql
	return indexer
}

func NewMultiUserIndexer(client *elastic.Client, session *r.Session, userIDs ...interface{}) *Indexer {
	rql := r.Table("users").GetAll(userIDs...)
	indexer := defaultUsersIndexer(client, session)
	indexer.RQL = rql
	return indexer
}

func defaultUsersIndexer(client *elastic.Client, session *r.Session) *Indexer {
	return &Indexer{
		GetID: func(item interface{}) string {
			user := item.(*schema.User)
			return user.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}
