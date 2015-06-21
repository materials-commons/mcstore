package mcstore

import (
	"time"

	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/uploads"
	"github.com/willf/bitset"
)

var _ = fmt.Println

// An uploadResource handles all upload requests.
type uploadResource struct {
	log *app.Logger
}

// UploadEntry is a client side representation of an upload.
type UploadEntry struct {
	RequestID   string    `json:"request_id"`
	FileName    string    `json:"filename"`
	DirectoryID string    `json:"directory_id"`
	ProjectID   string    `json:"project_id"`
	Size        int64     `json:"size"`
	Host        string    `json:"host"`
	Checksum    string    `json:"checksum"`
	Birthtime   time.Time `json:"birthtime"`
}

// newUploadResource creates a new upload resource
func newUploadResource() rest.Service {
	return &uploadResource{
		log: app.NewLog("resource", "upload"),
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

	ws.Route(ws.POST("/chunk").To(rest.RouteHandler(r.uploadFileChunk)).
		Consumes("multipart/form-data").
		Writes(UploadChunkResponse{}).
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
	ChunkSize     int32  `json:"chunk_size"`
	FileMTime     string `json:"filemtime"`
	Checksum      string `json: "checksum"`
}

// uploadCreateResponse is the format of JSON sent back containing
// the upload request ID.
type CreateUploadResponse struct {
	RequestID     string `json:"request_id"`
	StartingBlock uint   `json:"starting_block"`
}

type uploadRequester struct {
	idService  uploads.IDService
	dirService DirService
}

func newUploadRequester(session *r.Session) *uploadRequester {
	return &uploadRequester{
		idService:  uploads.NewIDServiceUsingSession(session),
		dirService: newDirServiceUsingSession(session),
	}
}

// createUploadRequest services requests to create a new upload id. It validates
// the given request, and ensures that the returned upload id is unique. Upload
// requests are persisted until deleted or a successful upload occurs.
func (ur *uploadResource) createUploadRequest(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*r.Session)
	uploadRequester := newUploadRequester(session)
	cr, err := uploadRequester.request2IDRequest(request, user.ID)
	if err != nil {
		app.Log.Debugf("request2IDRequst failed", err)
		return nil, err
	}

	upload, err := uploadRequester.idService.ID(cr)
	if err != nil {
		app.Log.Debugf("idService.ID failed", err)
		return nil, err
	}

	startingBlock := findStartingBlock(upload.File.Blocks)

	resp := CreateUploadResponse{
		RequestID:     upload.ID,
		StartingBlock: startingBlock,
	}

	return &resp, nil
}

// request2IDRequest fills out an id request to send to the idService. It handles request parameter errors.
func (u *uploadRequester) request2IDRequest(request *restful.Request, userID string) (uploads.IDRequest, error) {
	var req CreateUploadRequest
	var cr uploads.IDRequest

	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("request2IDRequest ReadEntity failed: %s", err)
		return cr, err
	}

	app.Log.Debugf("CreateUploadRequest %#v", req)

	fileMTime, err := time.Parse(time.RFC1123, req.FileMTime)
	if err != nil {
		app.Log.Debugf("request2IDRequest time.Parse failed on %s: %s", req.FileMTime, err)
		return cr, err
	}

	if req.ChunkSize == 0 {
		req.ChunkSize = 1024 * 1024
	}

	directoryID, err := u.getDirectoryID(req)
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
		ChunkSize:   req.ChunkSize,
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
func (u *uploadRequester) getDirectoryID(req CreateUploadRequest) (directoryID string, err error) {
	switch {
	case req.DirectoryID == "" && req.DirectoryPath == "":
		app.Log.Debugf("No directoryID or directoryPath specified")
		return "", app.ErrInvalid
	case req.DirectoryID != "":
		return req.DirectoryID, nil
	default:
		dir, err := u.dirService.createDir(req.ProjectID, req.DirectoryPath)
		if err != nil {
			app.Log.Debugf("CreateDir %s %s failed: %s", req.ProjectID, req.DirectoryPath, err)
			return "", err
		}
		return dir.ID, nil
	}
}

// findStartingBlock returns the block to start at. Blocks start
// at 1, since this is what the flow.js client expects. This
// method takes that into account.
func findStartingBlock(blocks *bitset.BitSet) uint {
	if blocks.None() {
		// Nothing set, start at 1
		return 1
	}

	// Create the complement and return first set.
	complement := blocks.Complement()
	if block, status := complement.NextSet(0); !status {
		// This shouldn't happen, but safest case is to check
		// for it and tell the client to resend everything.
		return 1
	} else {
		return block + 1
	}
}

type UploadChunkResponse struct {
	FileID string `json:"file_id"`
	Done   bool   `json:"done"`
}

// uploadFileChunk uploads a new file chunk.
func (ur *uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*r.Session)
	flowRequest, err := form2FlowRequest(request)
	if err != nil {
		ur.log.Errorf("Error converting form to flow.Request: %s", err)
		return nil, err
	}

	req := uploads.UploadRequest{
		Request: flowRequest,
	}

	uploadService := uploads.NewUploadServiceUsingSession(session)
	if uploadStatus, err := uploadService.Upload(&req); err != nil {
		return nil, err
	} else {
		uploadResp := &UploadChunkResponse{
			FileID: uploadStatus.FileID,
			Done:   uploadStatus.Done,
		}
		return uploadResp, nil
	}
}

// deleteUploadRequest will delete an existing upload request. It validates that
// the requesting user has access to delete the request.
func (ur *uploadResource) deleteUploadRequest(request *restful.Request, response *restful.Response, user schema.User) error {
	session := request.Attribute("session").(*r.Session)
	idService := uploads.NewIDServiceUsingSession(session)
	uploadID := request.PathParameter("id")
	return idService.Delete(uploadID, user.ID)
}

// listProjectUploadRequests returns the upload requests for the project if the requester
// has access to the project.
func (ur *uploadResource) listProjectUploadRequests(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*r.Session)
	idService := uploads.NewIDServiceUsingSession(session)
	projectID := request.PathParameter("project")
	entries, err := idService.UploadsForProject(projectID, user.ID)
	switch {
	case err == app.ErrNotFound:
		var uploads []UploadEntry
		return uploads, nil
	case err != nil:
		return nil, err
	default:
		return uploads2uploadEntries(entries), nil
	}
}

// uploads2uploadEntries converts schema.Upload array into an array of UploadEntry.
func uploads2uploadEntries(projectUploads []schema.Upload) []UploadEntry {
	entries := make([]UploadEntry, len(projectUploads))
	for i, entry := range projectUploads {
		entries[i] = UploadEntry{
			RequestID:   entry.ID,
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
