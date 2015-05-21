package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
)

// An projectsResource holds the state and services needed for the
// projects REST resource.
type projectsResource struct {
	log            *app.Logger
	dirService     DirService
	projectService ProjectService
}

//////////////////////// Request/Response Definitions /////////////////////

// createProjectRequest requests that a project be created. If MustNotExist
// is true, then the given project must not already exist. Existence is
// determined by the project name for that user.
type createProjectRequest struct {
	Name         string `json:"name"`
	MustNotExist bool   `json:"must_not_exist"`
}

// createProjectResponse returns the created project. If the project was an
// existing project and no new project was created then the Existing flag
// will be set to false.
type createProjectResponse struct {
	ProjectID string `json:"project_id"`
	Existing  bool   `json:"existing"`
}

// getDirectoryRequest is a request to get a directory for a project. The
// directory lookup is by path within the context of the given project.
type getDirectoryRequest struct {
	Path      string
	ProjectID string
}

// getDirectoryResponse returns the directory id for a directory
// path for a given project.
type getDirectoryResponse struct {
	DirectoryID string `json:"directory_id"`
	Path        string `json:"path"`
}

// newProjectsResource creates a new projects resource.
func newProjectsResource(dirService DirService, projectService ProjectService) *projectsResource {
	return &projectsResource{
		log:            app.NewLog("resource", "projects"),
		dirService:     dirService,
		projectService: projectService,
	}
}

// WebService creates an instance of the projects web service.
func (r *projectsResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/projects").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("").To(rest.RouteHandler(r.createProject)).
		Doc("Creates a new project for user. If project exists it returns the existing project.").
		Reads(createProjectRequest{}).
		Writes(createProjectResponse{}))

	ws.Route(ws.POST("directory").To(rest.RouteHandler(r.getDirectory)).
		Doc("Gets or creates a directory by its directory path").
		Reads(getDirectoryRequest{}).
		Writes(getDirectoryResponse{}))

	return ws
}

// createProject services the create project request. It will ensure that the user
// doesn't have a project matching the given project name.
func (r *projectsResource) createProject(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var req createProjectRequest
	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("createProject ReadEntity failed: %s", err)
		return nil, err
	}

	proj, existing, err := r.projectService.CreateProject(req.Name, user.ID, req.MustNotExist)
	switch {
	case err != nil:
		return nil, err
	default:
		resp := &createProjectResponse{
			ProjectID: proj.ID,
			Existing:  existing,
		}
		return resp, nil
	}
}

// getDirectory services request to get a directory for a project. It accepts directories
// by their path relative to the project. The getDirectory service will create a directory
// that doesn't exist.
func (r *projectsResource) getDirectory(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var req getDirectoryRequest
	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("getDirectory ReadEntity failed: %s", err)
		return nil, err
	}

	dir, err := r.dirService.CreateDir(req.ProjectID, req.Path)
	switch {
	case err != nil:
		return nil, err
	default:
		resp := &getDirectoryResponse{
			DirectoryID: dir.ID,
			Path:        req.Path,
		}
		return resp, nil
	}
}
