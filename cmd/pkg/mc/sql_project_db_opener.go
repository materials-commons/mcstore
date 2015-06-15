package mc

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"path/filepath"
)

type sqlProjectDBOpener struct {
	configer Configer
}

var ProjectOpener sqlProjectDBOpener = sqlProjectDBOpener{
	configer: NewOSUserConfiger(),
}

func (p sqlProjectDBOpener) OpenProjectDB(dbSpec ProjectDBSpec, flags ProjectOpenFlags) (ProjectDB, error) {
	switch flags {
	case ProjectDBCreate:
		return p.createProjectDB(dbSpec)
	case ProjectDBMustExist:
		return p.openProjectDB(dbSpec)
	default:
		return nil, app.ErrInvalid
	}
}

func (p sqlProjectDBOpener) createProjectDB(dbSpec ProjectDBSpec) (*sqlProjectDB, error) {
	dirPath := p.configer.ConfigDir()
	dbFilePath := filepath.Join(dirPath, dbSpec.ProjectID+".db")

	if err := validateDBPaths(dirPath, dbFilePath); err != nil {
		return nil, err
	}

	if db, err := p.openSqlDB(dbFilePath); err != nil {
		return nil, err
	} else {
		return p.loadDB(db, dbSpec)
	}
}

func validateDBPaths(dirPath, dbFilePath string) error {
	switch {
	case !file.Exists(dirPath):
		return app.ErrNotFound
	case file.Exists(dbFilePath):
		return app.ErrExists
	default:
		return nil
	}
}

// openDB will attempt to open the project.db file. The mustExist
// flag specifies whether or not the database file must exist.
func (p sqlProjectDBOpener) openSqlDB(dbpath string) (*sqlx.DB, error) {
	dbargs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", dbpath)
	db, err := sqlx.Open("sqlite3", dbargs)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (p sqlProjectDBOpener) loadDB(db *sqlx.DB, dbSpec ProjectDBSpec) (*sqlProjectDB, error) {
	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	proj := &Project{
		ProjectID: dbSpec.ProjectID,
		Name:      dbSpec.Name,
	}

	projectdb := &sqlProjectDB{
		db: db,
	}

	if _, err := projectdb.insertProject(proj); err != nil {
		db.Close()
		return nil, err
	}

	return projectdb, nil
}

func (p sqlProjectDBOpener) openProjectDB(dbSpec ProjectDBSpec) (*sqlProjectDB, error) {
	dirPath := p.configer.ConfigDir()
	dbFilePath := filepath.Join(dirPath, dbSpec.ProjectID+".db")

	if !file.Exists(dbFilePath) {
		return nil, app.ErrNotFound
	}

	if db, err := p.openSqlDB(dbFilePath); err != nil {
		return nil, err
	} else {
		proj := &sqlProjectDB{
			db: db,
		}
		return proj, nil
	}
}
