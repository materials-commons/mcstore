package upload

import (
	"time"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/server/mcstored/service/uploads"
)

// uploadCreateRequest describes the JSON request a client will send
// to create a new upload request.
type uploadCreateRequest struct {
	projectID   string `json:"project_id"`
	directoryID string `json:"directory_id"`
	userID      string `json:"user_id"`
}

// uploadCreateResponse is the format of JSON sent back containing
// the upload request ID.
type uploadCreateResponse struct {
	requestID string `json:"request_id"`
}

// createUploadRequest services requests to create a new upload id. It validates
// the given request, and ensures that the returned upload id is unique. Upload
// requests are persisted until deleted or a successful upload occurs.
func (r *uploadResource) createUploadRequest(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var req uploadCreateRequest
	if err := request.ReadEntity(&req); err != nil {
		return nil, err
	}

	cr := uploads.CreateRequest{
		User:        req.userID,
		DirectoryID: req.directoryID,
		ProjectID:   req.projectID,
		Host:        request.Request.RemoteAddr,
		Birthtime:   time.Now(),
	}
	upload, err := r.createService.Create(cr)
	if err != nil {
		return nil, err
	}

	resp := uploadCreateResponse{
		requestID: upload.ID,
	}
	return &resp, nil
}
