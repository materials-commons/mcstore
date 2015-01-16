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

// ByChecksum looks up a file by its checksum. This routine only returns the original
// root entry, it will not return entries that are duplicates and point at the root.
func (f rFiles) ByChecksum(checksum string) (*schema.File, error) {
	rql := model.Files.T().GetAllByIndex("checksum", checksum).Filter(r.Row.Field("usesid").Eq(""))
	var file schema.File
	if err := model.Files.Qs(f.session).Row(rql, &file); err != nil {
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

	// We will return one of the errors, even if both inserts failed. There really isn't anything the
	// system can do about a specific error, so we just want to communicate that a failure occured.
	err := model.DirFiles.Qs(f.session).Insert(&dir2file, nil)
	if err != nil {
		app.Log.Error(app.Logf("Unable to update datadir2datafile for file %s, directory %s, error %s", fileID, dirID, err))
	}

	err = model.ProjectFiles.Qs(f.session).Insert(&proj2file, nil)
	if err != nil {
		app.Log.Error(app.Logf("Unable to update project2datafile for file %s, project %s, error %s", fileID, projectID, err))
	}

	return err
}

// Update updates an existing datafile.
func (f rFiles) Update(file *schema.File) error {
	if err := model.Files.Qs(f.session).Update(file.ID, file); err != nil {
		return err
	}
	return nil
}

// UpdateFields updates the fields for the given id
func (f rFiles) UpdateFields(fileID string, fields map[string]interface{}) error {
	if err := model.Files.Qs(f.session).Update(fileID, fields); err != nil {
		return err
	}
	return nil
}
