package db

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/config"
)

type SessionCreater interface {
	RSession() (*r.Session, error)
	RSessionMust() *r.Session
}

type sessionCreater struct{}

var Sessions SessionCreater = &sessionCreater{}

func (_ *sessionCreater) RSession() (*r.Session, error) {
	return RSession()
}

func (_ *sessionCreater) RSessionMust() *r.Session {
	return RSessionMust()
}

// RSession creates a new RethinkDB session.
func RSession() (*r.Session, error) {
	return r.Connect(
		r.ConnectOpts{
			Address:  config.GetString("MCDB_CONNECTION"),
			Database: config.GetString("MCDB_NAME"),
		})
}

// RSessionMust creates a new RethinkDB session and panics if it cannot
// allocate it.
func RSessionMust() *r.Session {
	session, err := RSession()
	if err != nil {
		panic("Couldn't get new rethinkdb session")
	}
	return session
}

// RSessionUsing create a new RethinkDB session using the passed in parameters
func RSessionUsing(address, db string) (*r.Session, error) {
	return r.Connect(
		r.ConnectOpts{
			Address:  address,
			Database: db,
		})
}

// RSessionUsingMust creates a new RethinkDB session and panics if it cannot
// allocate it.
func RSessionUsingMust(address, db string) *r.Session {
	session, err := RSessionUsing(address, db)
	if err != nil {
		panic("Couldn't get new rethinkdb session")
	}
	return session
}
