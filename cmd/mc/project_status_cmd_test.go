package mc

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/server/mcstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	"net/http/httptest"
	"time"
)

var _ = fmt.Println

var _ = Describe("ProjectStatusCmd", func() {
	Describe("getUploads method tests", func() {
		var (
			client        *gorequest.SuperAgent
			server        *httptest.Server
			container     *restful.Container
			rr            *httptest.ResponseRecorder
			uploadRequest mcstore.CreateUploadRequest
		)

		BeforeEach(func() {
			client = newGoRequest()
			uploadRequest = mcstore.CreateUploadRequest{
				ProjectID:     "test",
				DirectoryID:   "test",
				DirectoryPath: "test/test",
				FileName:      "testreq.txt",
				FileSize:      4,
				FileMTime:     time.Now().Format(time.RFC1123),
				Checksum:      "abc123",
			}
			container = mcstore.NewServicesContainerForTest()
			server = httptest.NewServer(container)
			rr = httptest.NewRecorder()
			config.Set("mcurl", server.URL)
		})

		AfterEach(func() {
			server.Close()
		})

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

		It("Should return an error when the apikey doesn't exist", func() {
			config.Set("apikey", "does-not-exist")
			r, _, errs := client.Post(app.MCApi.APIUrl("/upload")).Send(uploadRequest).End()
			err := app.MCApi.APIError(r, errs)
			Expect(err).NotTo(BeNil())
			Expect(r.StatusCode).To(BeNumerically("==", 401))
			fmt.Println("r.StatusCode =", r.StatusCode)
		})

		It("Should get the upload entries for the project", func() {

		})
	})

	It("Should do something", func() {
		Expect("").To(Equal(""))
	})
})
