package upload

import (
	"time"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/server/mcstored/service/uploads"
)

type uploadCreateRequest struct {
	projectID   string `json:"project_id"`
	directoryID string `json:"directory_id"`
	userID      string `json:"user_id"`
}

type uploadCreateResponse struct {
	requestID string `json:"request_id"`
}

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
