package mcstore

import (
	"crypto/tls"

	"path"

	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/parnurzeal/gorequest"
	"gnd.la/net/urlutil"
)

type ServerAPI struct {
	agent  *gorequest.SuperAgent
	client *ezhttp.EzClient
}

// NewServerAPI creates a new ServerAPI
func NewServerAPI() *ServerAPI {
	return &ServerAPI{
		agent:  gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
		client: MCClient(),
	}
}

// CreateUploadRequest will request an upload request from the server. If an existing
// upload matches the request then server will send the existing upload request.
func (s *ServerAPI) CreateUploadRequest(req CreateUploadRequest) (*CreateUploadResponse, error) {
	r, body, errs := s.agent.Post(Url("/upload")).Send(req).End()
	if err := ToError(r, errs); err != nil {
		return nil, err
	}

	var uploadResponse CreateUploadResponse
	if err := ToJSON(body, &uploadResponse); err != nil {
		return nil, err
	}
	return &uploadResponse, nil
}

// SendFlowData will send the data for a flow request.
func (s *ServerAPI) SendFlowData(req *flow.Request) error {
	params := req.ToParamsMap()
	sc, err := s.client.PostFileBytes(Url("/upload/chunk"), "/tmp/test.txt", "chunkData",
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
	r, body, errs := s.agent.Get(Url("/upload/" + projectID)).End()
	if err := ToError(r, errs); err != nil {
		return nil, err
	}
	var entries []UploadEntry
	err := ToJSON(body, &entries)
	return entries, err
}

// DeleteUploadRequest will delete a given upload request.
func (s *ServerAPI) DeleteUploadRequest(uploadID string) error {
	r, _, errs := s.agent.Delete(Url("/upload/" + uploadID)).End()
	return ToError(r, errs)
}

// This really doesn't belong here as the server code is in a different server. However
// it logically belongs here as far as the client is concerned.

// userLogin contains the user password used to retrieve the users apikey.
type userLogin struct {
	Password string `json:"password"`
}

// GetUserAPIKey will return the users APIKey
func (s *ServerAPI) GetUserAPIKey(username, password string) (apikey string, err error) {
	l := userLogin{
		Password: password,
	}
	apiURL := urlutil.MustJoin(MCUrl(), path.Join("api", "user", username, "apikey"))
	r, body, errs := s.agent.Put(apiURL).Send(l).End()
	if err := ToError(r, errs); err != nil {
		return apikey, err
	}

	var u schema.User
	err = ToJSON(body, &u)
	return u.APIKey, err
}
