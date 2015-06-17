package mcstore

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
)

// ProjectService is a service for manipulating projects.
type ProjectService interface {
	createProject(projectName, owner string, mustNotExist bool) (*schema.Project, bool, error)
	getProjectByName(projectName, owner, user string) (*schema.Project, error)
	getProjectByID(projectID, user string) (*schema.Project, error)
}

// projectService implements the ProjectService interface
type projectService struct {
	projects dai.Projects
	dirs     dai.Dirs
	access   domain.Access
}

// newProjectService creates a new instance of the ProjectService. NewProjectService will panic
// if it cannot attach to the database.
func newProjectService() *projectService {
	session := db.RSessionMust()
	access := domain.NewAccess(dai.NewRProjects(session), dai.NewRFiles(session), dai.NewRUsers(session))

	return &projectService{
		projects: dai.NewRProjects(session),
		dirs:     dai.NewRDirs(session),
		access:   access,
	}
}

// newProjectServiceUsingSession creates a new idService that connects to the database using
// the given session.
func newProjectServiceUsingSession(session *r.Session) *projectService {
	return &projectService{
		projects: dai.NewRProjects(session),
		dirs:     dai.NewRDirs(session),
	}
}

// createProject will create a project. If mustNotExist is false, CreateProject will return the
// the project if it already exists. If mustNotExist is true and a matching project is found then
// CreateProject will return ErrExists, and the project will be nil. If CreateProject returns an
// existing project it will return true, otherwise it will return false.
func (s *projectService) createProject(projectName, owner string, mustNotExist bool) (*schema.Project, bool, error) {
	project, err := s.projects.ByName(projectName, owner)
	switch {
	case err != nil:
		return nil, false, err

	case project == nil:
		proj, err := s.createNewProject(projectName, owner)
		return proj, false, err

	default:
		if mustNotExist {
			return nil, true, app.ErrExists
		}

		return project, true, nil
	}
}

// createNewProject will create a new project.
func (s *projectService) createNewProject(projectName, owner string) (*schema.Project, error) {
	project := schema.NewProject(projectName, owner)
	return s.projects.Insert(&project)
}

// getProjectByName returns the project matching name owned by user. It validates
// access to the project and returns app.ErrNoAccess if user doesn't have access.
func (s *projectService) getProjectByName(projectName, owner, user string) (*schema.Project, error) {
	proj, err := s.projects.ByName(projectName, owner)
	switch {
	case err != nil:
		return nil, err
	case !s.access.AllowedByOwner(proj.ID, user):
		return nil, app.ErrNoAccess
	default:
		return proj, nil
	}
}

// getProjectByID returns the project matching id. It validates access to the project
// and returns app.ErrNoAccess if user doesn't have access.
func (s *projectService) getProjectByID(projectID, user string) (*schema.Project, error) {
	proj, err := s.projects.ByID(projectID)
	switch {
	case err != nil:
		return nil, err
	case !s.access.AllowedByOwner(proj.ID, user):
		return nil, app.ErrNoAccess
	default:
		return proj, nil
	}
}
