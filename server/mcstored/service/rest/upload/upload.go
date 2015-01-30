package upload

import (
	"github.com/emicklei/go-restful"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstored/service/uploads"
)

// An uploadResource handles all upload requests.
type uploadResource struct {
	log           log15.Logger
	idService     uploads.IDService
	uploadService uploads.UploadService
}

// NewResources creates a new upload resource
func NewResource(uploadService uploads.UploadService, idService uploads.IDService) rest.Service {
	return &uploadResource{
		log:           app.NewLog("resource", "upload"),
		idService:     idService,
		uploadService: uploadService,
	}
}

// WebService creates an instance of the upload web service.
func (r *uploadResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/upload").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)
	ws.Route(ws.POST("").To(rest.RouteHandler(r.createUploadRequest)).
		Doc("Creates a new upload request").
		Reads(uploadCreateRequest{}).
		Writes(uploadCreateResponse{}))
	ws.Route(ws.POST("/chunk").To(rest.RouteHandler1(r.uploadFileChunk)).
		Consumes("multipart/form-data").
		Doc("Upload a file chunk"))

	return ws
}

// uploadFileChunk uploads a new file chunk.
func (r *uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response, user schema.User) error {
	flowRequest, err := form2FlowRequest(request)
	if err != nil {
		r.log.Error(app.Logf("Error converting form to flow.Request: %s", err))
		return err
	}

	req := uploads.UploadRequest{
		Request: flowRequest,
	}
	return r.uploadService.Upload(&req)
}
