package model

import r "github.com/dancannon/gorethink"

type Model interface {
	Q() Query
	Qs(session *r.Session) Query
	Table() r.Term
	T() r.Term
}

type Query interface {
	Rql() r.Term
	Session() *r.Session
	ByID(id string, obj interface{}) error
	Row(query r.Term, obj interface{}) error
	Rows(query r.Term, results interface{}) error
	Update(id string, what interface{}) error
	InsertRaw(table string, what interface{}, dest interface{}) error
	Insert(what interface{}, dest interface{}) error
	Delete(id string) error
}
