package mcstore

import (
	"crypto/tls"

	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/parnurzeal/gorequest"
)

type ServerAPI struct {
	agent  *gorequest.SuperAgent
	client *ezhttp.EzClient
}

// NewServerAPI creates a new ServerAPI
func NewServerAPI() *ServerAPI {
	return &ServerAPI{
		agent:  gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
		client: Api.MCClient(),
	}
}

// CreateUploadRequest will request an upload request from the server. If an existing
// upload matches the request then server will send the existing upload request.
func (s *ServerAPI) CreateUploadRequest(req CreateUploadRequest) (*CreateUploadResponse, error) {
	r, body, errs := s.agent.Post(Api.Url("/upload")).Send(req).End()
	if err := Api.IsError(r, errs); err != nil {
		return nil, err
	}

	var uploadResponse CreateUploadResponse
	if err := Api.ToJSON(body, &uploadResponse); err != nil {
		return nil, err
	}
	return &uploadResponse, nil
}

// SendFlowData will send the data for a flow request.
func (s *ServerAPI) SendFlowData(req *flow.Request) error {
	params := req.ToParamsMap()
	sc, err := s.client.PostFileBytes(Api.Url("/upload/chunk"), "/tmp/test.txt", "chunkData",
		req.Chunk, params)
	switch {
	case err != nil:
		return err
	case sc != 200:
		return app.ErrInternal
	default:
		return nil
	}
}

// ListUploadRequests will return all the upload requests for a given project ID.
func (s *ServerAPI) ListUploadRequests(projectID string) ([]UploadEntry, error) {
	r, body, errs := s.agent.Get(Api.Url("/upload/" + projectID)).End()
	if err := Api.IsError(r, errs); err != nil {
		return nil, err
	}
	var entries []UploadEntry
	err := Api.ToJSON(body, &entries)
	return entries, err
}

// DeleteUploadRequest will delete a given upload request.
func (s *ServerAPI) DeleteUploadRequest(uploadID string) error {
	r, _, errs := s.agent.Delete(Api.Url("/upload/" + uploadID)).End()
	return Api.IsError(r, errs)
}
