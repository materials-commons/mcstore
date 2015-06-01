package mcstore

import (
	"time"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/uploads"
)

// An uploadResource handles all upload requests.
type uploadResource struct {
	log           *app.Logger
	idService     uploads.IDService
	uploadService uploads.UploadService
	dirService    DirService
}

// UploadEntry is a client side representation of an upload.
type UploadEntry struct {
	FileName    string    `json:"filename"`
	DirectoryID string    `json:"directory_id"`
	ProjectID   string    `json:"project_id"`
	Size        int64     `json:"size"`
	Host        string    `json:"host"`
	Checksum    string    `json:"checksum"`
	Birthtime   time.Time `json:"birthtime"`
}

// newUploadResource creates a new upload resource
func newUploadResource(uploadService uploads.UploadService, idService uploads.IDService, dirService DirService) rest.Service {
	return &uploadResource{
		log:           app.NewLog("resource", "upload"),
		idService:     idService,
		uploadService: uploadService,
		dirService:    dirService,
	}
}

// WebService creates an instance of the upload web service.
func (r *uploadResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/upload").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("").To(rest.RouteHandler(r.createUploadRequest)).
		Doc("Creates a new upload request").
		Reads(CreateUploadRequest{}).
		Writes(CreateUploadResponse{}))

	ws.Route(ws.POST("/chunk").To(rest.RouteHandler1(r.uploadFileChunk)).
		Consumes("multipart/form-data").
		Doc("Upload a file chunk"))

	ws.Route(ws.DELETE("{id}").To(rest.RouteHandler1(r.deleteUploadRequest)).
		Doc("Deletes an existing upload request").
		Param(ws.PathParameter("id", "upload request to delete").DataType("string")))

	ws.Route(ws.GET("{project}").To(rest.RouteHandler(r.listProjectUploadRequests)).
		Param(ws.PathParameter("project", "project id").DataType("string")).
		Doc("Lists upload requests for project").
		Writes([]UploadEntry{}))

	return ws
}

// CreateRequest describes the JSON request a client will send
// to create a new upload request.
type CreateUploadRequest struct {
	ProjectID     string `json:"project_id"`
	DirectoryID   string `json:"directory_id"`
	DirectoryPath string `json:"directory_path"`
	FileName      string `json:"filename"`
	FileSize      int64  `json:"filesize"`
	FileMTime     string `json:"filemtime"`
	Checksum      string `json: "checksum"`
}

// uploadCreateResponse is the format of JSON sent back containing
// the upload request ID.
type CreateUploadResponse struct {
	RequestID string `json:"request_id"`
}

// createUploadRequest services requests to create a new upload id. It validates
// the given request, and ensures that the returned upload id is unique. Upload
// requests are persisted until deleted or a successful upload occurs.
func (r *uploadResource) createUploadRequest(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	cr, err := r.request2IDRequest(request, user.ID)
	if err != nil {
		return nil, err
	}

	upload, err := r.idService.ID(cr)
	if err != nil {
		return nil, err
	}

	resp := CreateUploadResponse{
		RequestID: upload.ID,
	}

	return &resp, nil
}

// request2IDRequest fills out an id request to send to the idService. It handles request parameter errors.
func (r *uploadResource) request2IDRequest(request *restful.Request, userID string) (uploads.IDRequest, error) {
	var req CreateUploadRequest
	var cr uploads.IDRequest

	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("makeIDRequest ReadEntity failed: %s", err)
		return cr, err
	}

	app.Log.Debugf("%#v", req)

	fileMTime, err := time.Parse(time.RFC1123, req.FileMTime)
	if err != nil {
		app.Log.Debugf("makeIDRequest time.Parse failed on %s: %s", req.FileMTime, err)
		return cr, err
	}

	directoryID, err := r.getDirectoryID(req)
	if err != nil {
		app.Log.Debugf("makeIDRequest getDirectoryID failed: %s", err)
		return cr, err
	}

	cr = uploads.IDRequest{
		User:        userID,
		DirectoryID: directoryID,
		ProjectID:   req.ProjectID,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileMTime:   fileMTime,
		Checksum:    req.Checksum,
		Host:        request.Request.RemoteAddr,
		Birthtime:   time.Now(),
	}

	return cr, nil

}

// getDirectoryID returns the directoryID. A user can pass either a directoryID
// or a directory path. If a directory path is passed in, then the method will
// get the directoryID associated with that path in the project. If the path
// doesn't exist it will create it.
func (r *uploadResource) getDirectoryID(req CreateUploadRequest) (directoryID string, err error) {
	switch {
	case req.DirectoryID == "" && req.DirectoryPath == "":
		app.Log.Debugf("No directoryID or directoryPath specified")
		return "", app.ErrInvalid
	case req.DirectoryID != "":
		return req.DirectoryID, nil
	default:
		dir, err := r.dirService.createDir(req.ProjectID, req.DirectoryPath)
		if err != nil {
			app.Log.Debugf("CreateDir %s %s failed: %s", req.ProjectID, req.DirectoryPath, err)
			return "", err
		}
		return dir.ID, nil
	}
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
	return r.idService.Delete(uploadID, user.ID)
}

// listProjectUploadRequests returns the upload requests for the project if the requester
// has access to the project.
func (r *uploadResource) listProjectUploadRequests(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	projectID := request.PathParameter("project")
	entries, err := r.idService.UploadsForProject(projectID, user.ID)
	if err != nil {
		return nil, err
	}
	return uploads2uploadEntries(entries), nil
}

// uploads2uploadEntries converts schema.Upload array into an array of UploadEntry.
func uploads2uploadEntries(projectUploads []schema.Upload) []UploadEntry {
	entries := make([]UploadEntry, len(projectUploads))
	for i, entry := range projectUploads {
		entries[i] = UploadEntry{
			FileName:    entry.File.Name,
			DirectoryID: entry.DirectoryID,
			ProjectID:   entry.ProjectID,
			Size:        entry.File.Size,
			Host:        entry.Host,
			Checksum:    entry.File.Checksum,
			Birthtime:   entry.Birthtime,
		}
	}
	return entries
}
