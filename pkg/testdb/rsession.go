package testdb

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db"
)

var session *r.Session

func RSession() *r.Session {
	if session == nil {
		session = db.RSessionUsingMust("localhost:30815", "mctestdb")
	}
	return session
}

// RSessionErr always returns a nil err. It will panic if it cannot
// get a db session. This function is meant to be used with the
// databaseSessionFilter for unit testing.
func RSessionErr() (*r.Session, error) {
	return RSession(), nil
}
