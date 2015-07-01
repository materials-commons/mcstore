package mc

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/mcstore/pkg/app"
	_ "github.com/mattn/go-sqlite3"
)

// MCProject is a materials commons project on the local system.
type sqlProjectDB struct {
	db *sqlx.DB
}

// TODO: Remove this reference, only here to get project building during refactor.
func Find(dir string) (ProjectDB, error) {
	return nil, app.ErrInvalid
}

func (p *sqlProjectDB) Project() *Project {
	var proj Project
	if err := p.db.Get(&proj, "select * from project"); err != nil {
		app.Log.Critf("Unable to retrieve project from database: %s", err)
		return nil
	}
	return &proj
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

func (p *sqlProjectDB) UpdateDirectory(dir *Directory) error {
	sql := `
	       update directories set
	           directoryid = :directoryid,
	           path = :path,
	           lastupload = :lastupload,
	           lastdownload = :lastdownload
	       where id=:id
	`
	_, err := p.db.Exec(sql, dir.DirectoryID, dir.Path, dir.LastUpload, dir.LastDownload, dir.ID)
	return err
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

func (p *sqlProjectDB) FindFile(fileName string, dirID int64) (*File, error) {
	query := `select * from files where directory = $1 and name = $2`
	var f File
	err := p.db.Get(&f, query, dirID, fileName)
	switch {
	case err == sql.ErrNoRows:
		return nil, app.ErrNotFound
	case err != nil:
		return nil, err
	default:
		return &f, nil
	}
}

func (p *sqlProjectDB) UpdateFile(f *File) error {
	sql := `
	       update files set
	           fileid = :fileid,
	           name = :name,
	           checksum = :checksum,
	           size = :size,
	           mtime = :mtime,
	           ctime = :ctime,
	           lastupload = :lastupload,
	           lastdownload = :lastdownload,
	           directory = :directory
	       where id=:id
	`
	_, err := p.db.Exec(sql, f.FileID, f.Name, f.Checksum, f.Size, f.MTime,
		f.CTime, f.LastUpload, f.LastDownload, f.Directory, f.ID)
	return err
}

// insertProject will insert a new project entry into the project database.
func (p *sqlProjectDB) insertProject(proj *Project) (*Project, error) {
	sql := `
           insert into project(name, projectid, path, lastupload, lastdownload)
                       values(:name, :projectid, :path, :lastupload, :lastdownload)
        `
	res, err := p.db.Exec(sql, proj.Name, proj.ProjectID, proj.Path, proj.LastUpload, proj.LastDownload)
	if err != nil {
		return nil, err
	}
	proj.ID, _ = res.LastInsertId()
	return proj, err
}

func (p *sqlProjectDB) UpdateProject(proj *Project) error {
	sql := `
	     update project set
	         name = :name,
	         projectid = :projectid,
	         path = :path,
	         lastupload = :lastupload,
	         lastdownload = :lastdownload
	     where id=:id
	`
	_, err := p.db.Exec(sql, proj.Name, proj.ProjectID, proj.Path, proj.LastUpload, proj.LastDownload, proj.ID)
	return err
}

func (p *sqlProjectDB) Directories() []Directory {
	var dirs []Directory
	return dirs
}

func (p *sqlProjectDB) Ls(dir Directory) []File {
	var files []File
	return files
}

func (p *sqlProjectDB) Clone() ProjectDB {
	return p
}
