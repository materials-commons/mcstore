package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	_ "github.com/mattn/go-sqlite3"
)

type MCProject struct {
	ID  string
	Dir string
	db  *sqlx.DB
}

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

func Create(path, name, projectID string) (*MCProject, error) {
	projPath := filepath.Join(path, ".mcproject")
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
		Dir: filepath.Dir(path),
	}
	proj := &Project{
		ProjectID: projectID,
		Name:      name,
	}
	proj, err = mcproject.InsertProject(proj)
	if err != nil {
		db.Close()
		return nil, err
	}
	return mcproject, nil
}

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

func (p *MCProject) InsertProject(proj *Project) (*Project, error) {
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
