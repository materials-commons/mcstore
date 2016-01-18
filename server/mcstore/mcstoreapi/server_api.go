package mcstoreapi

import (
	"crypto/tls"

	"path"

	"path/filepath"
	"strings"

	"io"
	"net/http"
	"os"

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
func (s *ServerAPI) CreateUpload(req CreateUploadRequest) (*CreateUploadResponse, error) {
	var uploadResponse CreateUploadResponse
	sc, err := s.client.JSON(&req).JSONPost(Url("/upload"), &uploadResponse)
	if err != nil {
		return nil, err
	}
	err = HTTPStatusToError(sc)
	if err != nil {
		return nil, err
	}
	return &uploadResponse, nil
}

// SendFlowData will send the data for a flow request.
func (s *ServerAPI) SendFlowData(req *flow.Request) (*UploadChunkResponse, error) {
	params := req.ToParamsMap()
	sc, err, body := s.client.PostFileBytes(Url("/upload/chunk"), "/tmp/test.txt", "chunkData",
		req.Chunk, params)
	switch {
	case err != nil:
		return nil, err
	case sc != 200:
		return nil, app.ErrInternal
	default:
		var uploadResp UploadChunkResponse
		if err := ToJSON(body, &uploadResp); err != nil {
			return nil, err
		}
		return &uploadResp, nil
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

type DirectoryRequest struct {
	ProjectName string
	ProjectID   string
	Path        string
}

func (s *ServerAPI) GetDirectory(req DirectoryRequest) (directoryID string, err error) {
	var projectBasedPath string
	if projectBasedPath, err = toProjectPath(req.ProjectName, req.Path); err != nil {
		return directoryID, err
	}

	getDirReq := GetDirectoryRequest{
		Path:      projectBasedPath,
		ProjectID: req.ProjectID,
	}
	r, body, errs := s.agent.Post(Url("/project2/directory")).Send(getDirReq).End()
	if err = ToError(r, errs); err != nil {
		return directoryID, err
	}

	var dirResponse GetDirectoryResponse
	if err = ToJSON(body, &dirResponse); err != nil {
		return directoryID, err
	}

	return dirResponse.DirectoryID, nil
}

func (s *ServerAPI) GetDirectoryList(projectID, directoryID string) ([]string, error) {
	if directoryID == "" {
		directoryID = "top"
	}

	apiURL := urlutil.MustJoin(MCUrl(), path.Join("api", "projects", projectID, "directories", directoryID))
	if sc, err := s.client.JSONGet(Url(apiURL), nil); err != nil {
		return nil, err
	} else if err = HTTPStatusToError(sc); err != nil {
		return nil, err
	}

	return nil, nil
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

func (s *ServerAPI) DownloadFile(projectID, fileID, fpath string) error {
	out, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer out.Close()

	fileURL := Url(filepath.Join("/datafiles/static", fileID)) + "&original=true"
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (s *ServerAPI) GetFileForPath(projectID, fpath string) (*schema.File, error) {
	filePathArg := struct {
		FilePath string `json:"string"`
	}{
		FilePath: fpath,
	}

	apiURL := urlutil.MustJoin(MCUrl(), path.Join("api", "v2", "projects", projectID, "files_by_path"))
	r, body, errs := s.agent.Put(apiURL).Send(filePathArg).End()
	if err := ToError(r, errs); err != nil {
		return nil, err
	}

	var f schema.File
	if err := ToJSON(body, &f); err != nil {
		return nil, err
	}

	return &f, nil
}
