package projects

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstored/service/data"
)

// An projectsResource holds the state and services needed for the
// projects REST resource.
type projectsResource struct {
	log        *app.Logger
	dirService data.DirService
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

// NewResource creates a new projects resource.
func NewResource(dirService data.DirService) *projectsResource {
	return &projectsResource{
		log:        app.NewLog("resource", "projects"),
		dirService: dirService,
	}
}

// WebService creates an instance of the projects web service.
func (r *projectsResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/projects").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("directory").To(rest.RouteHandler(r.getDirectory)).
		Doc("Gets or creates a directory by its directory path").
		Reads(getDirectoryRequest{}).
		Writes(getDirectoryResponse{}))

	return ws
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
