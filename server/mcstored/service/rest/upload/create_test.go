package upload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/materials-commons/mcstore/server/mcstored/service/uploads"
	"github.com/materials-commons/mcstore/test"
	"github.com/stretchr/testify/require"
)

var (
	users    = dai.NewRUsers(test.RSession())
	files    = dai.NewRFiles(test.RSession())
	groups   = dai.NewRGroups(test.RSession())
	dirs     = dai.NewRDirs(test.RSession())
	projects = dai.NewRProjects(test.RSession())
	tuploads = dai.NewRUploads(test.RSession())
	access   = domain.NewAccess(groups, files, users)
)

func TestCreateUploadRequest(t *testing.T) {
	createService := uploads.NewCreateServiceFrom(dirs, projects, tuploads, access)
	rw := newSeparateItemRequestWriter()
	tracker := NewUploadTracker()
	uploader := NewUploader(rw, tracker)
	var b bytes.Buffer
	dest := bufio.NewWriter(&b)
	finisher := newTrackFinisher()
	af := newSeparateItemAssemblerFactory(rw, dest, finisher)
	uploadResource := NewResource(uploader, af, createService)
	container := restful.NewContainer()
	container.Add(uploadResource.WebService())
	ts := httptest.NewServer(container)
	defer ts.Close()
	url := ts.URL + "/upload"

	// Test with valid data
	jsonStr := []byte(`{
		"project_id": "test",
		"directory_id": "test",
		"filename": "test.txt",
		"filesize": 50,
		"filectime": "Tue, 18 Nov 2014 16:26:40 GMT",
		"user_id": "test@mc.org"
	}`)
	s := bytes.NewBuffer(jsonStr)
	req, err := http.NewRequest("POST", url, s)
	require.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	require.Nil(t, err)
	data, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	var v map[string]interface{}
	json.Unmarshal(data, &v)
	requestID := v["request_id"].(string)
	upload, err := tuploads.ByID(requestID)
	require.Nil(t, err)
	require.NotNil(t, upload)
	tuploads.Delete(requestID)

	// Test with invalid data
	jsonStr = []byte(`{
                "project_id": "no-project",
		"directory_id": "test",
		"filename": "test.txt",
		"filesize": 50,
		"filectime": "Tue, 18 Nov 2014 16:26:40 GMT",
		"user_id": "test@mc.org"
        }`)
	s = bytes.NewBuffer(jsonStr)
	req, err = http.NewRequest("POST", url, s)
	require.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 400)
}
