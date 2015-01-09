package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// rFiles implements the Files interface for RethinkDB
type rFiles struct {
	session *r.Session
}

// newRFiles creates a new instance of rFiles
func NewRFiles(session *r.Session) rFiles {
	return rFiles{
		session: session,
	}
}

// ByID looks up a file by its primary key. In RethinkDB this is the id field.
func (f rFiles) ByID(id string) (schema.File, error) {
	var file schema.File
	if err := model.Files.Qs(f.session).ByID(id, &file); err != nil {
		return nil, err
	}
	return file, nil
}
