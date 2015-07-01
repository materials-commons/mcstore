package mc

import (
	"path/filepath"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
)

type mcprojects struct {
	configer Configer
	dbOpener ProjectDBOpener
}

var Projects *mcprojects = NewProjects(NewOSUserConfiger())

func NewProjects(configer Configer) *mcprojects {
	return &mcprojects{
		configer: configer,
		dbOpener: sqlProjectDBOpener{configer: configer},
	}
}

func (p *mcprojects) All() ([]ProjectDB, error) {
	if !file.Exists(p.configer.ConfigDir()) {
		return nil, app.ErrNotFound
	}

	projectsGlob := filepath.Join(p.configer.ConfigDir(), "*.db")
	if fileMatches, err := filepath.Glob(projectsGlob); err != nil {
		return nil, err
	} else {
		return p.loadProjectDBEntries(fileMatches), nil
	}
}

func (p *mcprojects) loadProjectDBEntries(projectDBPaths []string) []ProjectDB {
	var projects []ProjectDB
	for _, filePath := range projectDBPaths {
		name := p.dbOpener.PathToName(filePath)
		if projdb, err := p.dbOpener.OpenProjectDB(name); err != nil {
			app.Log.Errorf("Unable to open projectDB '%s': %s", filePath, err)
		} else {
			projects = append(projects, projdb)
		}
	}
	return projects
}

func (p *mcprojects) Create(dbSpec ProjectDBSpec) (ProjectDB, error) {
	return p.dbOpener.CreateProjectDB(dbSpec)
}
