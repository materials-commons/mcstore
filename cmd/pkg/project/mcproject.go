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
	dbpath := filepath.Join(dir, "project.db")
	dbargs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", dbpath)
	db, err := sqlx.Open("sqlite3", dbargs)
	if err != nil {
		return nil, err
	}

	mcproject := &MCProject{
		db: db,
	}
	return mcproject, app.ErrInvalid
}
