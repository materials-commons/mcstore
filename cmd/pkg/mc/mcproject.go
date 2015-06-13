package mc

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	_ "github.com/mattn/go-sqlite3"
)

// MCProject is a materials commons project on the local system.
type MCProject struct {
	ID  string
	Dir string
	db  *sqlx.DB
}

// Clone will create a new instance of MCProject from and existing one. This can
// be used to create separate MCProjects for different go routines.
func (p *MCProject) Clone() *MCProject {
	mcprojectDir := filepath.Join(p.Dir, ".mcproject")
	db, _ := openDB(mcprojectDir, true)
	return &MCProject{
		ID:  p.ID,
		Dir: p.Dir,
		db:  db,
	}
}

// Find will attempt to locate the project that a directory is in. It does
// this by going up the directory tree looking for a .mcproject directory
// that contains a project.db in it. If it cannot find one it returns
// the error app.ErrNotFound
func Find(dir string) (*MCProject, error) {
	// Normalize the directory path, and convert all path separators to a
	// forward slash (/).
	dirPath, err := filepath.Abs(dir)
	if err != nil {
		app.Log.Debugf("Bad directory %s: %s", dir, err)
		return nil, err
	}

	dirPath = filepath.ToSlash(dirPath)
	for {
		if dirPath == "/" {
			// Projects at root level not allowed
			return nil, app.ErrNotFound
		}

		mcprojectDir := filepath.Join(dirPath, ".mcproject")
		if file.IsDir(mcprojectDir) {
			return Open(mcprojectDir)
		}
		dirPath = filepath.Dir(dirPath)
	}
}

// Open will open a project database in the specified .mcproject
// directory. The project.db must exist. If it doesn't exist
// it will return the error app.ErrNotFound.
func Open(dir string) (*MCProject, error) {
	db, err := openDB(dir, true)
	if err != nil {
		return nil, err
	}

	mcproject := &MCProject{
		db: db,
		// Dir of location.
		Dir: filepath.Dir(dir),
	}
	return mcproject, nil
}

// openDB will attempt to open the project.db file. The mustExist
// flag specifies whether or not the database file must exist.
func openDB(dir string, mustExist bool) (*sqlx.DB, error) {
	dbpath := filepath.Join(dir, "project.db")
	if mustExist && !file.Exists(dbpath) {
		return nil, app.ErrNotFound
	}
	dbargs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", dbpath)
	db, err := sqlx.Open("sqlite3", dbargs)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type ClientProject struct {
	Name      string
	ProjectID string
	Path      string
}

// Create will create a new .mcproject directory in path and
// populate the database with the given project.
func Create(project ClientProject) (*MCProject, error) {
	projPath := filepath.Join(project.Path, ".mcproject")
	if err := os.MkdirAll(projPath, 0700); err != nil {
		return nil, err
	}

	db, err := openDB(projPath, false)
	if err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	mcproject := &MCProject{
		db: db,
		// Dir of location.
		Dir: filepath.Dir(project.Path),
	}
	proj := &Project{
		ProjectID: project.ProjectID,
		Name:      project.Name,
	}
	proj, err = mcproject.insertProject(proj)
	if err != nil {
		db.Close()
		return nil, err
	}
	return mcproject, nil
}

// InsertDirectory will insert a new directory entry into the project database.
func (p *MCProject) InsertDirectory(dir *Directory) (*Directory, error) {
	sql := `
           insert into directories(directoryid, path, lastupload, lastdownload)
                       values(:directoryid, :path, :lastupload, :lastdownload)
        `
	res, err := p.db.Exec(sql, dir.DirectoryID, dir.Path, dir.LastUpload, dir.LastDownload)
	if err != nil {
		return nil, err
	}
	dir.ID, _ = res.LastInsertId()
	return dir, nil
}

// FindDirectoryByPath looks up a directory by its path.
func (p *MCProject) FindDirectoryByPath(path string) (*Directory, error) {
	query := `select * from directories where path = $1`
	var dir Directory

	err := p.db.Get(&dir, query, path)
	switch {
	case err == sql.ErrNoRows:
		return nil, app.ErrNotFound
	case err != nil:
		return nil, err
	default:
		return &dir, nil
	}
}

// InsertFile will insert a new file entry into the project database.
func (p *MCProject) InsertFile(f *File) (*File, error) {
	sql := `
           insert into files(fileid, name, checksum, size, mtime,
                             ctime, lastupload, lastdownload, directory)
                       values(:fileid, :name, :checksum, :size, :mtime,
                              :ctime, :lastupload, :lastdownload, :directory)
        `
	res, err := p.db.Exec(sql, f.FileID, f.Name, f.Checksum, f.Size, f.MTime,
		f.CTime, f.LastUpload, f.LastDownload, f.Directory)
	if err != nil {
		return nil, err
	}
	f.ID, _ = res.LastInsertId()
	return f, err
}

// insertProject will insert a new project entry into the project database.
func (p *MCProject) insertProject(proj *Project) (*Project, error) {
	sql := `
           insert into project(name, projectid, lastupload, lastdownload)
                       values(:name, :projectid, :lastupload, :lastdownload)
        `
	res, err := p.db.Exec(sql, proj.Name, proj.ProjectID, proj.LastUpload, proj.LastDownload)
	if err != nil {
		return nil, err
	}
	proj.ID, _ = res.LastInsertId()
	return proj, err
}
