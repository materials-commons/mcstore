package mc

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	_ "github.com/mattn/go-sqlite3"
)

// MCProject is a materials commons project on the local system.
type sqlProjectDB struct {
	db *sqlx.DB
}

func Open(dbpath string) (*sqlProjectDB, error) {
	if !file.Exists(dbpath) {
		return nil, app.ErrNotFound
	}

	if db, err := openDB(dbpath); err != nil {
		return nil, err
	} else {
		proj := &sqlProjectDB{
			db: db,
		}
		return proj, nil
	}
}

// TODO: Remove this reference, only here to get project building during refactor.
func Find(dir string) (ProjectDB, error) {
	return nil, app.ErrInvalid
}

// openDB will attempt to open the project.db file. The mustExist
// flag specifies whether or not the database file must exist.
func openDB(dbpath string) (*sqlx.DB, error) {
	dbargs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", dbpath)
	db, err := sqlx.Open("sqlite3", dbargs)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type ProjectReq struct {
	Name      string
	ProjectID string
	Path      string
}

// Create will create a new project database at the named path.
// It will return an error if the project already exists.
func Create(projectReq ProjectReq, path string) (*sqlProjectDB, error) {
	dbfilePath := filepath.Join(path, projectReq.ProjectID+".db")

	switch {
	case !file.Exists(path):
		return nil, app.ErrNotFound
	case file.Exists(dbfilePath):
		return nil, app.ErrExists
	}

	db, err := openDB(dbfilePath)
	if err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	mcproject := &sqlProjectDB{
		db: db,
	}
	proj := &Project{
		ProjectID: projectReq.ProjectID,
		Name:      projectReq.Name,
	}
	proj, err = mcproject.insertProject(proj)
	if err != nil {
		db.Close()
		return nil, err
	}
	return mcproject, nil
}

func (p *sqlProjectDB) Project() *Project {
	return nil
}

// InsertDirectory will insert a new directory entry into the project database.
func (p *sqlProjectDB) InsertDirectory(dir *Directory) (*Directory, error) {
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

// FindDirectory looks up a directory by its path.
func (p *sqlProjectDB) FindDirectory(path string) (*Directory, error) {
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
func (p *sqlProjectDB) InsertFile(f *File) (*File, error) {
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
func (p *sqlProjectDB) insertProject(proj *Project) (*Project, error) {
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

func (p *sqlProjectDB) Directories() []Directory {
	var dirs []Directory
	return dirs
}

func (p *sqlProjectDB) Ls(dir Directory) []File {
	var files []File
	return files
}
