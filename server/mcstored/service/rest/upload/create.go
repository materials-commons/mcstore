package upload

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/schema"
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

	return nil, nil
}
