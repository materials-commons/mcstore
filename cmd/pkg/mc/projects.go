package mc

import (
	"path/filepath"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
)

type mcprojects struct {
	config   Configer
	dbOpener ProjectDBOpener
}

var Projects *mcprojects = NewProjects(NewOSUserConfiger())

func NewProjects(config Configer) *mcprojects {
	return &mcprojects{
		config: config,
		// TODO: Remove hard coding of type of opener here
		dbOpener: sqlProjectDBOpener{},
	}
}

func (p *mcprojects) All() ([]ProjectDB, error) {
	if !file.Exists(p.config.ConfigDir()) {
		return nil, app.ErrNotFound
	}

	projectsGlob := filepath.Join(p.config.ConfigDir(), "*.db")
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

func (p *mcprojects) Create(project *Project) (ProjectDB, error) {
	return nil, app.ErrInvalid
}
