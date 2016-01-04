package mcstore

import (
	"fmt"

	rethinkdb "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
)

// An projectsResource holds the state and services needed for the
// projects REST resource.
type projectsResource struct {
	log *app.Logger
}

// newProjectsResource creates a new projects resource.
func newProjectsResource() *projectsResource {
	return &projectsResource{
		log: app.NewLog("resource", "projects"),
	}
}

// WebService creates an instance of the projects web service.
func (r *projectsResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/project2").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("").To(rest.RouteHandler(r.createProject)).
		Doc("Creates a new project for user. If project exists it returns the existing project.").
		Reads(mcstoreapi.CreateProjectRequest{}).
		Writes(mcstoreapi.CreateProjectResponse{}))

	ws.Route(ws.POST("directory").To(rest.RouteHandler(r.getDirectory)).
		Doc("Gets or creates a directory by its directory path").
		Reads(mcstoreapi.GetDirectoryRequest{}).
		Writes(mcstoreapi.GetDirectoryResponse{}))

	ws.Route(ws.GET("{id}").To(rest.RouteHandler(r.getProject)).
		Doc("Gets project details").
		Param(ws.PathParameter("id", "project id").DataType("string")).
		Writes(ProjectEntry{}))

	ws.Route(ws.GET("").To(rest.RouteHandler(r.getUsersProjects)).
		Doc("Gets all projects user has access to").
		Writes([]ProjectEntry{}))

	return ws
}

// createProject services the create project request. It will ensure that the user
// doesn't have a project matching the given project name.
func (r *projectsResource) createProject(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*rethinkdb.Session)
	var req mcstoreapi.CreateProjectRequest
	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("createProject ReadEntity failed: %s", err)
		return nil, err
	}

	projectService := newProjectService(session)
	proj, existing, err := projectService.createProject(req.Name, user.ID, req.MustNotExist)
	switch {
	case err != nil:
		return nil, err
	default:
		resp := &mcstoreapi.CreateProjectResponse{
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
	fmt.Println("getDirectory found")
	session := request.Attribute("session").(*rethinkdb.Session)
	var req mcstoreapi.GetDirectoryRequest
	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("getDirectory ReadEntity failed: %s", err)
		return nil, err
	}

	dirService := newDirService(session)
	dir, err := dirService.createDir(req.ProjectID, req.Path)
	switch {
	case err != nil:
		return nil, err
	default:
		resp := &mcstoreapi.GetDirectoryResponse{
			DirectoryID: dir.ID,
			Path:        req.Path,
		}
		return resp, nil
	}
}

type ProjectEntry struct {
}

func (r *projectsResource) getProject(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	//	projectID := request.PathParameter("id")
	//	r.projectService.getProject(projectID, user.ID, false)
	return nil, nil
}

func (r *projectsResource) getUsersProjects(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	return nil, nil
}
