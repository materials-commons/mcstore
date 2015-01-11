package upload

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/emicklei/go-restful"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
)

const chunkPerms = 0700 // Permissions to set uploads to

// An uploadResource handles all upload requests.
type uploadResource struct {
	ctracker *chunkTracker
	log      log15.Logger // Resource specific logging.
}

// NewResources creates a new upload resource
func NewResource() rest.Service {
	return &uploadResource{
		ctracker: newChunkTracker(),
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

	// Ensure directory path exists
	uploadPath, err := r.createUploadDir(flowRequest)
	if err != nil {
		r.log.Error(app.Logf("Unable to create temporary chunk space: %s", err))
		return err
	}

	// Write chunk and determine if done.
	if err := r.processChunk(uploadPath, flowRequest); err != nil {
		r.log.Error(app.Logf("Unable to write chunk for file: %s", err))
		return err
	}

	return nil
}

// createUploadDir creates the directory for the chunk.
func (r *uploadResource) createUploadDir(flowRequest *flow.Request) (string, error) {
	uploadPath := fileUploadPath(flowRequest.ProjectID, flowRequest.DirectoryID, flowRequest.FileID)

	// os.MkdirAll returns nil if the path already exists.
	return uploadPath, os.MkdirAll(uploadPath, chunkPerms)
}

// processChunk writes the chunk and determines if this is the last chunk to write.
// If the last chunk has been uploaded it kicks off a reassembly of the file.
func (r *uploadResource) processChunk(uploadPath string, flowRequest *flow.Request) error {
	cpath := chunkPath(uploadPath, flowRequest.FlowChunkNumber)
	if err := r.writeChunk(cpath, flowRequest.Chunk); err != nil {
		return err
	}

	if r.uploadDone(flowRequest) {
		r.finishUpload(flowRequest)
	}

	return nil
}

// uploadDone checks to see if the upload has finished.
func (r *uploadResource) uploadDone(flowRequest *flow.Request) bool {
	id := flowRequest.UploadID()
	count := r.ctracker.addChunk(id)
	return count == flowRequest.FlowTotalChunks
}

// writeChunk writes a file chunk.
func (r *uploadResource) writeChunk(chunkpath string, chunk []byte) error {
	return ioutil.WriteFile(chunkpath, chunk, chunkPerms)
}

// finishUpload marks the upload as finished and kicks off an assembler to assemble the file.
func (r *uploadResource) finishUpload(flowRequest *flow.Request) {
	id := flowRequest.UploadID()
	r.ctracker.clear(id)
	assembler := newAssemberFromFlowRequest(flowRequest)
	go assembler.assembleFile()
}
