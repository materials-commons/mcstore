package projects

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstored/service/data"
)

type projectsResource struct {
	log        *app.Logger
	dirService data.DirService
}

type getDirectoryRequest struct {
	Path      string
	ProjectID string
}

type getDirectoryResponse struct {
	DirectoryID string `json:"directory_id"`
	Path        string `json:"path"`
}

func NewResource(dirService data.DirService) *projectsResource {
	return &projectsResource{
		log:        app.NewLog("resource", "projects"),
		dirService: dirService,
	}
}

func (r *projectsResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/projects").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("directory").To(rest.RouteHandler(r.getDirectory)).
		Doc("Gets or creates a directory by its directory path").
		Reads(getDirectoryRequest{}).
		Writes(getDirectoryResponse{}))

	return ws
}

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
