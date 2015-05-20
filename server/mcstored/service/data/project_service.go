package data

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// ProjectService is a service for manipulating projects.
type ProjectService interface {
	CreateProject(projectName, owner string, mustNotExist bool) (*schema.Project, bool, error)
	GetProject(projectName, owner string, create bool) (*schema.Project, bool, error)
}

// projectService implements the ProjectService interface
type projectService struct {
	projects dai.Projects
	dirs     dai.Dirs
}

// NewProjectService creates a new instance of the ProjectService. NewProjectService will panic
// if it cannot attach to the database.
func NewProjectService() *projectService {
	session := db.RSessionMust()
	return &projectService{
		projects: dai.NewRProjects(session),
		dirs:     dai.NewRDirs(session),
	}
}

// CreateProject will create a project. If mustNotExist is false, CreateProject will return the
// the project if it already exists. If mustNotExist is true and a matching project is found then
// CreateProject will return ErrExists, and the project will be nil. If CreateProject returns an
// existing project it will return true, otherwise it will return false.
func (s *projectService) CreateProject(projectName, owner string, mustNotExist bool) (*schema.Project, bool, error) {
	project, err := s.projects.ByName(projectName, owner)
	switch {
	case err != nil:
		return nil, false, err

	case project == nil:
		proj, err := s.createProject(projectName, owner)
		return proj, false, err

	default:
		if mustNotExist {
			return nil, false, app.ErrExists
		}

		return project, true, nil
	}
}

// GetProject attempts to get a project by project name for the given user. If the create flag is
// set then it will create the project if it doesn't exist. If GetProject creates a project it
// will return true, otherwise it will return false.
func (s *projectService) GetProject(projectName, owner string, create bool) (*schema.Project, bool, error) {
	project, err := s.projects.ByName(projectName, owner)

	switch {
	case err != nil:
		return nil, false, err

	case project == nil:
		if create {
			proj, err := s.createProject(projectName, owner)
			return proj, true, err
		}
		return nil, false, app.ErrNotFound

	default:
		return project, false, nil
	}
}

// createProject will create a new project.
func (s *projectService) createProject(projectName, owner string) (*schema.Project, error) {
	project := schema.NewProject(projectName, "", owner)
	return s.projects.Insert(&project)
}
