package project

import (
	"fmt"
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
	db, err := openDB(dir)
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

func openDB(dir string) (*sqlx.DB, error) {
	dbpath := filepath.Join(dir, "project.db")
	dbExists := file.Exists(dbpath)
	dbargs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", dbpath)
	db, err := sqlx.Open("sqlite3", dbargs)
	if err != nil {
		return nil, err
	}
	if !dbExists {
		if err := createSchema(db); err != nil {
			db.Close()
			return nil, err
		}
	}
	return db, nil
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
