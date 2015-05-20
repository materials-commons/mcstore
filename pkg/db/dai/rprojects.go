package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// rProjects implements the Projects interface for RethinkDB
type rProjects struct {
	session *r.Session
}

// NewRProjects creates a new instance of rProjects.
func NewRProjects(session *r.Session) rProjects {
	return rProjects{
		session: session,
	}
}

// ByID looks up a project by the given id.
func (p rProjects) ByID(id string) (*schema.Project, error) {
	var project schema.Project
	if err := model.Projects.Qs(p.session).ByID(id, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

func (p rProjects) ByName(name string, owner string) (*schema.Project, error) {
	var projects []schema.Project
	rql := r.Table("projects").GetAllByIndex("owner", owner).
		Filter(r.Row.Field("name").Eq(name))

	if err := model.Projects.Qs(p.session).Rows(rql, &projects); err != nil {
		return nil, err
	}

	numProjects := len(projects)
	switch {
	case numProjects == 0:
		return nil, nil
	case numProjects > 1:
		app.Log.Critf("Projects table corrupted, there are multiple projects for user '%s' with name '%s'", owner, name)
		return &projects[0], nil
	default:
		return &projects[0], nil
	}
}

// Insert creates a new project for the given owner. It also creates the directory associated
// with the project.
func (p rProjects) Insert(project *schema.Project) (*schema.Project, error) {
	if project.DataDir != "" {
		return nil, app.ErrInvalid
	}

	var (
		newProject schema.Project
		newDir     *schema.Directory
		err        error
	)

	if err = model.Projects.Qs(p.session).Insert(project, &newProject); err != nil {
		return nil, app.ErrCreate
	}

	dir := schema.NewDirectory(project.Name, project.Owner, newProject.ID, "")
	rdirs := NewRDirs(p.session)

	if newDir, err = rdirs.Insert(&dir); err != nil {
		return nil, app.ErrCreate
	}

	newProject.DataDir = newDir.ID
	if err = model.Projects.Qs(p.session).Update(newProject.ID, &newProject); err != nil {
		return &newProject, err
	}

	err = p.AddDirectories(&newProject, newDir.ID)

	return &newProject, err
}

// AddDirectories adds new directories to the project.
func (p rProjects) AddDirectories(project *schema.Project, directoryIDs ...string) error {
	var rverror error

	// Add each directory to the project2datadir table. If there are any errors,
	// remember that we saw an error, but continue on.
	for _, dirID := range directoryIDs {
		p2d := schema.Project2DataDir{
			ProjectID: project.ID,
			DataDirID: dirID,
		}
		if err := model.Projects.Qs(p.session).InsertRaw("project2datadir", p2d, nil); err != nil {
			rverror = app.ErrCreate
		}
	}

	return rverror
}

// HasDirectory checks if the given directoryID is in the given project.
func (p rProjects) HasDirectory(projectID, dirID string) bool {
	rql := model.ProjectDirs.T().GetAllByIndex("datadir_id", dirID)
	var proj2dir []schema.Project2DataDir
	if err := model.ProjectDirs.Qs(p.session).Rows(rql, &proj2dir); err != nil {
		return false
	}

	// Look for matching projectID
	for _, entry := range proj2dir {
		if entry.ProjectID == projectID {
			return true
		}
	}

	return false
}

// Get the access list for this project.
func (p rProjects) AccessList(projectID string) ([]schema.Access, error) {
	rql := r.Table("access").Filter(r.Row.Field("project_id").Eq(projectID))
	var access []schema.Access
	if err := model.Access.Qs(p.session).Rows(rql, &access); err != nil {
		return nil, err
	}
	return access, nil
}
