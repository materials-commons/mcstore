package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/app"
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

// ByPath looks up a directory in a project by its path.
func (d rDirs) ByPath(path, projectID string) (*schema.Directory, error) {
	rql := model.Dirs.T().GetAllByIndex("name", path).Filter(r.Row.Field("project").Eq(projectID))
	var dir schema.Directory
	if err := model.Dirs.Qs(d.session).Row(rql, &dir); err != nil {
		return nil, err
	}
	return &dir, nil
}

// Files returns the files for the given directory id.
func (d rDirs) Files(dirID string) ([]schema.File, error) {
	var files []schema.File
	rql := r.Table("datadir2datafile").GetAllByIndex("datadir_id", dirID).
		EqJoin("datafile_id", r.Table("datafiles")).Zip()
	if err := model.Files.Qs(d.session).Rows(rql, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// Insert creates a new dir.
func (d rDirs) Insert(dir *schema.Directory) (*schema.Directory, error) {
	var newDir schema.Directory
	if err := model.Dirs.Qs(d.session).Insert(dir, &newDir); err != nil {
		return nil, err
	}

	proj2dir := schema.Project2DataDir{
		ProjectID: dir.Project,
		DataDirID: newDir.ID,
	}
	if err := model.ProjectDirs.Qs(d.session).Insert(proj2dir, nil); err != nil {
		return nil, err
	}
	return &newDir, nil
}

// Delete will delete the directory from the project and the directory
// entry. If there are files, you should call the delete for files before
// deleting the directory.
func (d rDirs) Delete(dirID string) error {
	if err := model.Dirs.Qs(d.session).Delete(dirID); err != nil {
		return err
	}
	rql := model.ProjectDirs.T().GetAllByIndex("datadir_id", dirID).Delete()
	rv, err := rql.RunWrite(d.session)
	switch {
	case err != nil:
		return err
	case rv.Errors != 0:
		return app.ErrNotFound
	case rv.Deleted == 0:
		return app.ErrNotFound
	default:
		return nil
	}
}
