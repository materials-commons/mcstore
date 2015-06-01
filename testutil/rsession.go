package testutil

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db"
)

var _session *r.Session

func RSession() *r.Session {
	if _session == nil {
		_session = db.RSessionUsingMust("localhost:30815", "mctestdb")
	}
	return _session
}
