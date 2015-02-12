package projects

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
)

type projectsResource struct {
}

func NewResource() rest.Service {
	return &projectsResource{}
}

type projectTreeRequest struct{}

func (r *projectsResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/projects").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)
	ws.Route(ws.GET("/tree").To(rest.RouteHandler(r.projectTree)).
		Doc("Create the project tree").
		Reads(projectTreeRequest{}))
	return ws
}

func (r *projectsResource) projectTree(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	return nil, nil
}
