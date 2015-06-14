package mc

import (
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"path/filepath"
)

type mcprojects struct {
	config Configer
}

var Projects *mcprojects = NewProjects(NewOSUserConfiger())

func NewProjects(config Configer) *mcprojects {
	return &mcprojects{
		config: config,
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
		return loadProjectDBEntries(fileMatches), nil
	}
}

func loadProjectDBEntries(projectDBPaths []string) []ProjectDB {
	var projects []ProjectDB
	for _, filePath := range projectDBPaths {
		if projdb, err := Open(filePath); err != nil {
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
