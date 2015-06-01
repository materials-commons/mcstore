package testutil

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
