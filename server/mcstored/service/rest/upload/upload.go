package upload

import (
	"github.com/emicklei/go-restful"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
)

// An uploadResource handles all upload requests.
type uploadResource struct {
	uploader *uploader
	log      log15.Logger
	factory  AssemblerFactory
}

// NewResources creates a new upload resource
func NewResource(uploader *uploader, factory AssemblerFactory) rest.Service {
	return &uploadResource{
		uploader: uploader,
		log:      app.NewLog("resource", "upload"),
		factory:  factory,
	}
}

// WebService creates an instance of the upload web service.
func (r *uploadResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/upload").Produces(restful.MIME_JSON)
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

	if err := r.uploader.processRequest(flowRequest); err != nil {
		return err
	}

	if r.uploader.allBlocksUploaded(flowRequest) {
		go r.assembler(flowRequest)
	}

	return nil
}

// assembler builds a new Assembler to assemble the pieces of the file.
func (r *uploadResource) assembler(request *flow.Request) {
	if assembler := r.factory.Assembler(request); assembler != nil {
		assembler.Assemble()
	}
}
