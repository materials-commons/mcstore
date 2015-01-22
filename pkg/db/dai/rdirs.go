package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// rDirs implements the Dirs interface for RethinkDB.
type rDirs struct {
	session *r.Session
}

// NewRDirs creates a new instance of rDirs.
func NewRDirs(session *r.Session) rDirs {
	return rDirs{
		session: session,
	}
}

// ByID looks up a directory by the given id.
func (d rDirs) ByID(id string) (*schema.Directory, error) {
	var dir schema.Directory
	if err := model.Dirs.Qs(d.session).ByID(id, &dir); err != nil {
		return nil, err
	}
	return &dir, nil
}
