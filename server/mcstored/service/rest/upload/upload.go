package upload

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
)

const chunkPerms = 0700 // Permissions to set uploads to

// An uploadResource handles all upload requests.
type uploadResource struct {
	uploader *uploader
	log      log15.Logger // Resource specific logging.
}

// NewResources creates a new upload resource
func NewResource() rest.Service {
	return &uploadResource{
		uploader: newUploader(),
		log:      app.NewLog("resource", "upload"),
	}
}

// WebService creates an instance of the upload web service.
func (r *uploadResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/upload").
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/chunk").To(rest.RouteHandler1(r.uploadFileChunk)).
		Consumes("multipart/form-data").
		Doc("Upload a file chunk"))
	// ws.Route(ws.GET("/chunk").To(rest.RouteHandler(r.testFileChunk)).
	// 	Reads(flow.Request{}).
	// 	Doc("Test if chunk already uploaded."))

	return ws
}

// testFileChunk checks if a chunk has already been uploaded. At the moment we don't
// support this functionality, so always return an error.
func (r *uploadResource) testFileChunk(request *restful.Request, response *restful.Response) {
	response.WriteErrorString(http.StatusInternalServerError, "no such file")
}

// uploadFileChunk uploads a new file chunk.
func (r *uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response, user schema.User) error {
	// Create request
	flowRequest, err := form2FlowRequest(request)
	if err != nil {
		r.log.Error(app.Logf("Error converting form to flow.Request: %s", err))
		return err
	}

	return r.uploader.processRequest(flowRequest)
}
