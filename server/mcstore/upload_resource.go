package mcstore

import (
	"time"

	"fmt"

	rethinkdb "github.com/dancannon/gorethink"
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

	ws.Route(ws.POST("").Filter(projectAccessFilter).Filter(directoryFilter).To(rest.RouteHandler(r.createUploadRequest)).
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

	ws.Route(ws.GET("{project}").Filter(projectAccessFilter).To(rest.RouteHandler(r.listProjectUploadRequests)).
		Param(ws.PathParameter("project", "project id").DataType("string")).
		Doc("Lists upload requests for project").
		Writes([]UploadEntry{}))

	return ws
}

// CreateRequest describes the JSON request a client will send
// to create a new upload request.
type CreateUploadRequest struct {
	ProjectID   string `json:"project_id"`
	DirectoryID string `json:"directory_id"`
	FileName    string `json:"filename"`
	FileSize    int64  `json:"filesize"`
	ChunkSize   int32  `json:"chunk_size"`
	FileMTime   string `json:"filemtime"`
	Checksum    string `json: "checksum"`
}

// uploadCreateResponse is the format of JSON sent back containing
// the upload request ID.
type CreateUploadResponse struct {
	RequestID     string `json:"request_id"`
	StartingBlock uint   `json:"starting_block"`
}

// createUploadRequest services requests to create a new upload id. It validates
// the given request, and ensures that the returned upload id is unique. Upload
// requests are persisted until deleted or a successful upload occurs.
func (r *uploadResource) createUploadRequest(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	if cr, err := request2IDRequest(request, user.ID); err != nil {
		app.Log.Debugf("request2IDRequst failed", err)
		return nil, err
	} else {
		session := request.Attribute("session").(*rethinkdb.Session)
		project := request.Attribute("project").(schema.Project)
		directory := request.Attribute("directory").(schema.Directory)
		idService := uploads.NewIDService(session)

		if upload, err := idService.ID(cr, &project, &directory); err != nil {
			app.Log.Debugf("idService.ID failed", err)
			return nil, err
		} else {
			startingBlock := findStartingBlock(upload.File.Blocks)
			resp := CreateUploadResponse{
				RequestID:     upload.ID,
				StartingBlock: startingBlock,
			}
			return &resp, nil
		}
	}
}

// request2IDRequest fills out an id request to send to the idService. It handles request parameter errors.
func request2IDRequest(request *restful.Request, userID string) (uploads.IDRequest, error) {
	var (
		req CreateUploadRequest
		cr  uploads.IDRequest
	)

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

	cr = uploads.IDRequest{
		User:        userID,
		DirectoryID: req.DirectoryID,
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
func (r *uploadResource) uploadFileChunk(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*rethinkdb.Session)
	flowRequest, err := form2FlowRequest(request)
	if err != nil {
		r.log.Errorf("Error converting form to flow.Request: %s", err)
		return nil, err
	}

	req := uploads.UploadRequest{
		Request: flowRequest,
	}

	uploadService := uploads.NewUploadService(session)
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
func (r *uploadResource) deleteUploadRequest(request *restful.Request, response *restful.Response, user schema.User) error {
	session := request.Attribute("session").(*rethinkdb.Session)
	idService := uploads.NewIDService(session)
	uploadID := request.PathParameter("id")
	return idService.Delete(uploadID, user.ID)
}

// listProjectUploadRequests returns the upload requests for the project if the requester
// has access to the project.
func (r *uploadResource) listProjectUploadRequests(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*rethinkdb.Session)
	idService := uploads.NewIDService(session)
	project := request.Attribute("project").(schema.Project)
	entries, err := idService.UploadsForProject(project.ID)
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
