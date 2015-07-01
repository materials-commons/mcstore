package mcstore

import (
	"crypto/tls"

	"path"

	"path/filepath"
	"strings"

	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/gohandy/with"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/parnurzeal/gorequest"
	"gnd.la/net/urlutil"
)

type ServerAPI struct {
	agent  *gorequest.SuperAgent
	client *ezhttp.EzClient
	with.Retrier
}

// NewServerAPI creates a new ServerAPI
func NewServerAPI() *ServerAPI {
	retrier := with.NewRetrier()
	retrier.RetryCount = 0
	return &ServerAPI{
		agent:   gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
		client:  MCClient(),
		Retrier: retrier,
	}
}

func (s *ServerAPI) SetRetrier(r with.Retrier) {
	s.Retrier = r
}

// CreateUploadRequest will request an upload request from the server. If an existing
// upload matches the request then server will send the existing upload request.
func (s *ServerAPI) CreateUploadRequest(req CreateUploadRequest) (*CreateUploadResponse, error) {
	var (
		statusCode     int
		err            error
		uploadResponse CreateUploadResponse
	)
	fn := func() bool {
		if statusCode, err = s.client.JSON(&req).JSONPost(Url("/upload"), &uploadResponse); err != nil {
			return false
		} else {
			err = HTTPStatusToError(statusCode)
			return true
		}
	}
	s.WithRetry(fn)
	return &uploadResponse, nil
}

// SendFlowData will send the data for a flow request.
func (s *ServerAPI) SendFlowData(req *flow.Request) (*UploadChunkResponse, error) {
	var (
		body       string
		statusCode int
		err        error
		uploadResp UploadChunkResponse
	)
	fn := func() bool {
		params := req.ToParamsMap()
		statusCode, err, body = s.client.PostFileBytes(Url("/upload/chunk"), "/tmp/test.txt", "chunkData",
			req.Chunk, params)
		if err != nil {
			return false
		}
		switch {
		case statusCode != 200:
			err = app.ErrInternal
		default:
			var uploadResp UploadChunkResponse
			err = ToJSON(body, &uploadResp)
		}
		return true
	}

	s.WithRetry(fn)
	if err != nil {
		return nil, err
	}
	return &uploadResp, nil
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

type DirectoryRequest struct {
	ProjectName string
	ProjectID   string
	Path        string
}

func (s *ServerAPI) GetDirectory(req DirectoryRequest) (emptyDirectoryID string, err error) {
	var projectBasedPath string
	if projectBasedPath, err = toProjectPath(req.ProjectName, req.Path); err != nil {
		return emptyDirectoryID, err
	}

	getDirReq := GetDirectoryRequest{
		Path:      projectBasedPath,
		ProjectID: req.ProjectID,
	}
	r, body, errs := s.agent.Post(Url("/projects/directory")).Send(getDirReq).End()
	if err = ToError(r, errs); err != nil {
		return emptyDirectoryID, err
	}

	var dirResponse GetDirectoryResponse
	if err = ToJSON(body, &dirResponse); err != nil {
		return emptyDirectoryID, err
	}

	return dirResponse.DirectoryID, nil
}

func toProjectPath(projectName, path string) (string, error) {
	i := strings.Index(path, projectName)
	if i == -1 {
		return "", app.ErrInvalid
	}
	return filepath.ToSlash(path[i:]), nil
}

func (s *ServerAPI) CreateProject(req CreateProjectRequest) (*CreateProjectResponse, error) {
	var response CreateProjectResponse
	sc, err := s.client.JSON(&req).JSONPost(Url("/projects"), &response)
	if err != nil {
		return nil, err
	}

	err = HTTPStatusToError(sc)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
