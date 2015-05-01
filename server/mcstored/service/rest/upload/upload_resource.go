package upload

import (
	"time"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstored/service/uploads"
)

// An uploadResource handles all upload requests.
type uploadResource struct {
	log           *app.Logger
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
		Reads(CreateRequest{}).
		Writes(CreateResponse{}))
	ws.Route(ws.POST("/chunk").To(rest.RouteHandler1(r.uploadFileChunk)).
		Consumes("multipart/form-data").
		Doc("Upload a file chunk"))
	ws.Route(ws.DELETE("{id}").To(rest.RouteHandler1(r.deleteUploadRequest)).
		Doc("Deletes an existing upload request").
		Param(ws.PathParameter("id", "upload request to delete").DataType("string")))

	return ws
}

// CreateRequest describes the JSON request a client will send
// to create a new upload request.
type CreateRequest struct {
	ProjectID   string `json:"project_id"`
	DirectoryID string `json:"directory_id"`
	FileName    string `json:"filename"`
	FileSize    int64  `json:"filesize"`
	FileMTime   string `json:"filemtime"`
}

// uploadCreateResponse is the format of JSON sent back containing
// the upload request ID.
type CreateResponse struct {
	RequestID string `json:"request_id"`
}

// createUploadRequest services requests to create a new upload id. It validates
// the given request, and ensures that the returned upload id is unique. Upload
// requests are persisted until deleted or a successful upload occurs.
func (r *uploadResource) createUploadRequest(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var req CreateRequest
	if err := request.ReadEntity(&req); err != nil {
		return nil, err
	}

	fileMTime, err := time.Parse(time.RFC1123, req.FileMTime)
	if err != nil {
		return nil, err
	}

	cr := uploads.IDRequest{
		User:        user.ID,
		DirectoryID: req.DirectoryID,
		ProjectID:   req.ProjectID,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileMTime:   fileMTime,
		Host:        request.Request.RemoteAddr,
		Birthtime:   time.Now(),
	}
	upload, err := r.idService.ID(cr)
	if err != nil {
		return nil, err
	}

	resp := CreateResponse{
		RequestID: upload.ID,
	}
	return &resp, nil
}

// uploadFileChunk uploads a new file chunk.
func (r *uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response, user schema.User) error {
	flowRequest, err := form2FlowRequest(request)
	if err != nil {
		r.log.Errorf("Error converting form to flow.Request: %s", err)
		return err
	}

	req := uploads.UploadRequest{
		Request: flowRequest,
	}
	return r.uploadService.Upload(&req)
}

// deleteUploadRequest will delete an existing upload request. It validates that
// the requesting user has access to delete the request.
func (r *uploadResource) deleteUploadRequest(request *restful.Request, response *restful.Response, user schema.User) error {
	uploadID := request.PathParameter("id")
	if uploadID == "" {
		return app.ErrInvalid
	}

	return r.idService.Delete(uploadID, user.ID)
}
