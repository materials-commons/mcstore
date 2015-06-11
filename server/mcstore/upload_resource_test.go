package mcstore

import (
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/config"
	c "github.com/materials-commons/mcstore/cmd/pkg/client"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
)

var _ = fmt.Println

var _ = Describe("UploadResource", func() {
	var (
		client        *gorequest.SuperAgent
		server        *httptest.Server
		container     *restful.Container
		rr            *httptest.ResponseRecorder
		uploadRequest CreateUploadRequest
		uploads       dai.Uploads
	)

	BeforeEach(func() {
		client = c.NewGoRequest()
		container = NewServicesContainerForTest()
		server = httptest.NewServer(container)
		rr = httptest.NewRecorder()
		config.Set("mcurl", server.URL)
		uploadRequest = CreateUploadRequest{
			ProjectID:     "test",
			DirectoryID:   "test",
			DirectoryPath: "test/test",
			FileName:      "testreq.txt",
			FileSize:      4,
			FileMTime:     time.Now().Format(time.RFC1123),
			Checksum:      "abc123",
		}
		uploads = dai.NewRUploads(testutil.RSession())
	})

	var (
		createUploadRequest = func(req CreateUploadRequest) (*CreateUploadResponse, error) {
			r, body, errs := client.Post(app.MCApi.APIUrl("/upload")).Send(req).End()
			if err := app.MCApi.APIError(r, errs); err != nil {
				return nil, err
			}

			var uploadResponse CreateUploadResponse
			if err := app.MCApi.ToJSON(body, &uploadResponse); err != nil {
				return nil, err
			}
			return &uploadResponse, nil
		}
	)

	AfterEach(func() {
		server.Close()
	})

	Describe("create upload tests", func() {
		It("Should return an error when the user doesn't have permission", func() {
			// Set apikey for user who doesn't have permission
			config.Set("apikey", "test2")
			r, _, errs := client.Post(app.MCApi.APIUrl("/upload")).Send(uploadRequest).End()
			err := app.MCApi.APIError(r, errs)
			Expect(err).NotTo(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 401))
		})

		It("Should return an error when the project doesn't exist", func() {
			config.Set("apikey", "test")
			uploadRequest.ProjectID = "does-not-exist"
			r, _, errs := client.Post(app.MCApi.APIUrl("/upload")).Send(uploadRequest).End()
			err := app.MCApi.APIError(r, errs)
			Expect(err).NotTo(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 400))
		})

		It("Should return an error when the directory doesn't exist", func() {
			config.Set("apikey", "test")
			uploadRequest.DirectoryID = "does-not-exist"
			r, _, errs := client.Post(app.MCApi.APIUrl("/upload")).Send(uploadRequest).End()
			err := app.MCApi.APIError(r, errs)
			Expect(err).NotTo(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 400))
		})

		It("Should return an error when the apikey doesn't exist", func() {
			config.Set("apikey", "does-not-exist")
			r, _, errs := client.Post(app.MCApi.APIUrl("/upload")).Send(uploadRequest).End()
			err := app.MCApi.APIError(r, errs)
			Expect(err).NotTo(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 401))
		})

		It("Should create a new request for a valid submit", func() {
			config.Set("apikey", "test")
			r, body, errs := client.Post(app.MCApi.APIUrl("/upload")).Send(uploadRequest).End()
			err := app.MCApi.APIError(r, errs)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 200))
			var uploadResponse CreateUploadResponse
			err = app.MCApi.ToJSON(body, &uploadResponse)
			Expect(err).To(BeNil())

			uploadEntry, err := uploads.ByID(uploadResponse.RequestID)
			Expect(err).To(BeNil())
			Expect(uploadEntry.ID).To(Equal(uploadResponse.RequestID))
			err = uploads.Delete(uploadEntry.ID)
			Expect(err).To(BeNil())
		})
	})

	Describe("get uploads tests", func() {
		It("Should return an error on a bad apikey", func() {
			config.Set("apikey", "test")
			resp, err := createUploadRequest(uploadRequest)
			Expect(err).To(BeNil())

			config.Set("apikey", "bad-key")
			r, _, errs := client.Get(app.MCApi.APIUrl("/upload/test")).End()
			err = app.MCApi.APIError(r, errs)
			Expect(err).ToNot(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 401))

			err = uploads.Delete(resp.RequestID)
			Expect(err).To(BeNil())
		})

		It("Should return an error on a bad project", func() {
			config.Set("apikey", "test")
			r, _, errs := client.Get(app.MCApi.APIUrl("/upload/bad-project-id")).End()
			err := app.MCApi.APIError(r, errs)
			Expect(err).ToNot(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 400))
		})

		It("Should get existing upload requests for a project", func() {
			config.Set("apikey", "test")
			resp, err := createUploadRequest(uploadRequest)
			Expect(err).To(BeNil())
			r, body, errs := client.Get(app.MCApi.APIUrl("/upload/test")).End()
			err = app.MCApi.APIError(r, errs)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 200))
			var entries []UploadEntry
			err = app.MCApi.ToJSON(body, &entries)
			Expect(err).To(BeNil())
			Expect(len(entries)).To(BeNumerically("==", 1))
			entry := entries[0]
			Expect(entry.RequestID).To(Equal(resp.RequestID))

			err = uploads.Delete(resp.RequestID)
			Expect(err).To(BeNil())
		})
	})
})
