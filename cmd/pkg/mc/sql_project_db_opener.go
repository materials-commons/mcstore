package mc

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
)

type sqlProjectDBOpener struct {
	configer Configer
}

var ProjectOpener sqlProjectDBOpener = sqlProjectDBOpener{
	configer: NewOSUserConfiger(),
}

func (p sqlProjectDBOpener) CreateProjectDB(dbSpec ProjectDBSpec) (ProjectDB, error) {
	dirPath := p.configer.ConfigDir()
	dbFilePath := filepath.Join(dirPath, dbSpec.Name+".db")

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
		Path:      dbSpec.Path,
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

func (p sqlProjectDBOpener) OpenProjectDB(name string) (ProjectDB, error) {
	dirPath := p.configer.ConfigDir()
	dbFilePath := filepath.Join(dirPath, name+".db")

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

func (p sqlProjectDBOpener) PathToName(path string) string {
	dbfile := filepath.Base(path)
	lastDot := strings.LastIndex(dbfile, ".")

	// Remove extension if it exists
	if lastDot == -1 {
		return dbfile
	}
	return dbfile[:lastDot]
}
