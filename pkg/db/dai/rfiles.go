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

// NewRFiles creates a new instance of rFiles
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

// ByPath looks up a file by its name in a specific directory. It only returns the
// current file, not hidden files.
func (f rFiles) ByPath(name, dirID string) (*schema.File, error) {
	rql := r.Table("datadir2datafile").GetAllByIndex("datadir_id", dirID).
		EqJoin("datafile_id", r.Table("datafiles")).
		Zip().
		Filter(r.Row.Field("current").Eq(true).And(r.Row.Field("name").Eq(name)))
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

// Delete will delete the file from the directory and project. It will remove
// the file entry if the file isn't in any other directories or projects, or
// if it isn't referred to by other files (via usesid). If the file is no longer
// in a project and directory, but it is referenced by a usesid then it will remain
// in place as a disconnected file. In this case its current flag will be set to false.
// If you are deleting a file that has a parentid, then the parent will be set to
// current. This method will attempt to clean up as much as possible even in the face
// of errors. If there are any errors it will return the first error. It is the calling
// routines duty to figure out what steps could not be performed.
func (f rFiles) Delete(fileID, directoryID, projectID string) (*schema.File, error) {
	var firstError error // Keep the first error around

	// Get all projects this file is in
	projects, err := f.getProjects(fileID)
	if err != nil {
		firstError = err
	}

	// Get all directories this file is in
	dirs, err := f.getDirs(fileID)
	if err != nil && firstError == nil {
		firstError = err
	}

	// Delete file from directory
	err = f.deleteFromDir(fileID, directoryID)
	if err != nil && firstError == nil {
		firstError = err
	}

	file, err := f.ByID(fileID)
	if err != nil && firstError == nil {
		// Ugh, if we can't get the file then there
		// is nothing else we can do.
		return nil, err
	}

	filesUsedBy, err := f.getUsedBy(fileID)

	// We have deleted the file from one directory. At this
	// point if len(dirs) == 1 and len(filesUsedBy) == 0 and
	// len(projects) == 1 we can delete the entry.
	//
	// TODO: This logic is actually a lot more complicated when
	// files start being shared across projects and if we also
	// allow a file to be shared across directories in a project.
	// Right now across directories (in a project) isn't supported
	// but across projects is something we want to support.
	switch {
	case len(dirs) == 1 && len(projects) == 1 && len(filesUsedBy) == 0:
		model.Files.Qs(f.session).Delete(fileID)
	case len(projects) == 1 && len(dirs) == 1:
		// delete from project
		f.deleteFromProject(fileID, projectID)
	case file.Current:
		// File is referenced by somebody, so just mark it as
		// not current, since we cannot delete it.
		file.Current = false
		f.Update(file)
	}

	if file.Parent != "" {
		fields := map[string]interface{}{
			schema.FileFields.Current(): true,
		}
		f.UpdateFields(file.Parent, fields)
	}

	return file, firstError
}

// getProjects returns a list of all the projects containing this fileID.
func (f rFiles) getProjects(fileID string) ([]schema.Project2DataFile, error) {
	rql := model.ProjectFiles.T().GetAllByIndex("datafile_id", fileID)
	var projects []schema.Project2DataFile
	if err := model.ProjectFiles.Qs(f.session).Rows(rql, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// getDirs returns a list of all the directories containing this fileID.
func (f rFiles) getDirs(fileID string) ([]schema.DataDir2DataFile, error) {
	rql := model.DirFiles.T().GetAllByIndex("datafile_id", fileID)
	var dirs []schema.DataDir2DataFile
	if err := model.DirFiles.Qs(f.session).Rows(rql, &dirs); err != nil {
		return nil, err
	}
	return dirs, nil
}

// getUsedBy returns all the files that point at this file.
func (f rFiles) getUsedBy(fileID string) ([]schema.File, error) {
	rql := model.Files.T().GetAllByIndex("usesid", fileID)
	var files []schema.File
	if err := model.Files.Qs(f.session).Rows(rql, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// deleteFromDir will delete the given file from the directory.
func (f rFiles) deleteFromDir(fileID, directoryID string) error {
	rql := model.DirFiles.T().GetAllByIndex("datafile_id", fileID).
		Filter(r.Row.Field("datadir_id").Eq(directoryID)).Delete()
	rv, err := rql.RunWrite(f.session)
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

// deleteFromProject will delete the given file from the project.
func (f rFiles) deleteFromProject(fileID, projectID string) error {
	rql := model.ProjectFiles.T().GetAllByIndex("datafile_id", fileID).
		Filter(r.Row.Field("project_id").Eq(projectID)).Delete()
	rv, err := rql.RunWrite(f.session)
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
