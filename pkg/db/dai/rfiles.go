package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
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
func (f rFiles) ByID(id string) (*schema.File, error) {
	var file schema.File
	if err := model.Files.Qs(f.session).ByID(id, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// Insert adds a new file to the system.
func (f rFiles) Insert(file *schema.File, dirID string, projectID string) (*schema.File, error) {
	var newFile schema.File
	if err := model.Files.Qs(f.session).Insert(file, &newFile); err != nil {
		return nil, err
	}

	err := f.updateDependencies(newFile.ID, dirID, projectID)
	return &newFile, err
}

// updateDependencies updates the datadir2datafile and project2datafile tables.
func (f rFiles) updateDependencies(fileID, dirID, projectID string) error {
	proj2file := schema.Project2DataFile{
		ProjectID:  projectID,
		DataFileID: fileID,
	}

	dir2file := schema.DataDir2DataFile{
		DataDirID:  dirID,
		DataFileID: fileID,
	}

	err1 := model.Files.Qs(f.session).InsertRaw("datadir2datafile", &dir2file, nil)
	app.Log.Error(app.Logf("Unable to update datadir2datafile for file %s, directory %s, error %s", fileID, dirID, err1))

	err2 := model.Files.Qs(f.session).InsertRaw("project2datafile", &proj2file, nil)
	app.Log.Error(app.Logf("Unable to update project2datafile for file %s, project %s, error %s", fileID, projectID, err2))

	// We will return one of the errors, even if both inserts failed. There really isn't anything the
	// system can do about a specific error, so we just want to communicate that a failure occured.
	switch {
	case err1 != nil:
		return err1
	case err2 != nil:
		return err2
	default:
		return nil
	}
}
