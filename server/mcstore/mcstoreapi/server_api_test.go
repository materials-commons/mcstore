package mcstoreapi

import (
	"time"

	"net/http/httptest"

	"fmt"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/testdb"
	"github.com/materials-commons/mcstore/server/mcstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = fmt.Println

var _ = Describe("ServerAPI", func() {
	var (
		api           *ServerAPI
		server        *httptest.Server
		container     *restful.Container
		rr            *httptest.ResponseRecorder
		uploads       dai.Uploads
		uploadRequest CreateUploadRequest
	)

	BeforeEach(func() {
		container = mcstore.NewServicesContainer(testdb.Sessions)
		server = httptest.NewServer(container)
		rr = httptest.NewRecorder()
		config.Set("mcurl", server.URL)
		config.Set("apikey", "test")
		uploads = dai.NewRUploads(testdb.RSessionMust())
		api = NewServerAPI()
		uploadRequest = CreateUploadRequest{
			ProjectID:   "test",
			DirectoryID: "test",
			FileName:    "testreq.txt",
			FileSize:    4,
			ChunkSize:   2,
			FileMTime:   time.Now().Format(time.RFC1123),
			Checksum:    "abc12345",
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("CreateUploadRequest", func() {
		var resp *CreateUploadResponse
		var err error

		AfterEach(func() {
			if resp != nil {
				uploads.Delete(resp.RequestID)
			}
		})

		It("Should create an upload request", func() {
			resp, err = api.CreateUpload(uploadRequest)
			Expect(err).To(BeNil())
			Expect(resp.RequestID).NotTo(Equal(""))
			Expect(resp.StartingBlock).To(BeNumerically("==", 1))
		})

		It("Should return the same id for a duplicate upload request", func() {
			resp, err = api.CreateUpload(uploadRequest)
			Expect(err).To(BeNil())
			Expect(resp.RequestID).NotTo(Equal(""))
			Expect(resp.StartingBlock).To(BeNumerically("==", 1))

			resp2, err := api.CreateUpload(uploadRequest)
			Expect(err).To(BeNil())
			Expect(resp2.RequestID).To(Equal(resp.RequestID))
			Expect(resp.StartingBlock).To(BeNumerically("==", 1))
		})
	})

	Describe("SendFlowData", func() {
		var flowReq flow.Request
		var resp *CreateUploadResponse
		var err error

		BeforeEach(func() {
			flowReq = flow.Request{
				FlowChunkNumber:  1,
				FlowTotalChunks:  2,
				FlowChunkSize:    2,
				FlowTotalSize:    4,
				FlowFileName:     "testreq.txt",
				FlowRelativePath: "test/testreq.txt",
				ProjectID:        "test",
				DirectoryID:      "test",
			}
		})

		AfterEach(func() {
			if resp != nil {
				uploads.Delete(resp.RequestID)
			}
		})

		It("Should fail on an invalid request id", func() {
			flowReq.FlowIdentifier = "i-dont-exist"
			cresp, err := api.SendFlowData(&flowReq)
			Expect(err).To(Equal(app.ErrInvalid))
			Expect(cresp).To(BeNil())
		})

		It("Should Send the data an increment and increment starting block", func() {
			resp, err = CreateUploadRequest(uploadRequest)
			Expect(err).To(BeNil())
			flowReq.FlowIdentifier = resp.RequestID
			cresp, err := api.SendFlowData(&flowReq)
			Expect(err).To(BeNil())
			Expect(cresp.Done).To(BeFalse())

			resp2, err := api.CreateUpload(uploadRequest)
			Expect(err).To(BeNil())
			Expect(resp2.RequestID).To(Equal(resp.RequestID))
			Expect(resp2.StartingBlock).To(BeNumerically("==", 2))
		})
	})

	Describe("ListUploadRequests", func() {
		var resp *CreateUploadResponse

		AfterEach(func() {
			if resp != nil {
				uploads.Delete(resp.RequestID)
			}
		})

		It("Should return an empty list when there are no upload requests", func() {
			uploads, err := api.ListUploadRequests("test")
			Expect(err).To(BeNil())
			Expect(uploads).To(HaveLen(0))
		})

		It("Should return a list with one request when a single upload request has been created", func() {
			var err error
			resp, err = api.CreateUpload(uploadRequest)
			Expect(err).To(BeNil())
			uploads, err := api.ListUploadRequests("test")
			Expect(err).To(BeNil())
			Expect(uploads).To(HaveLen(1))
		})
	})

	Describe("DeleteUploadRequest", func() {
		var resp *CreateUploadResponse

		AfterEach(func() {
			if resp != nil {
				uploads.Delete(resp.RequestID)
			}
		})

		It("Should return an error if upload request doesn't exist", func() {
			err := api.DeleteUploadRequest("does-not-exist")
			Expect(err).NotTo(BeNil())
		})

		It("Should return an error if user doesn't have permission", func() {
			var err error
			resp, err = api.CreateUpload(uploadRequest)
			Expect(err).To(BeNil())

			// Change to a user who doesn't have permission
			config.Set("apikey", "test2")

			err = api.DeleteUploadRequest(resp.RequestID)
			Expect(err).NotTo(BeNil())
		})

		It("Should succeed if request exists and user has permission", func() {
			var err error
			resp, err = CreateUploadRequest(uploadRequest)
			Expect(err).To(BeNil())
			err = api.DeleteUploadRequest(resp.RequestID)
			Expect(err).To(BeNil())
		})
	})

	Describe("GetDirectory", func() {
		var (
			dirs       dai.Dirs = dai.NewRDirs(testdb.RSessionMust())
			dirID      string
			dirRequest DirectoryRequest
		)

		BeforeEach(func() {
			dirID = ""
			dirRequest = DirectoryRequest{
				ProjectName: "test",
				ProjectID:   "test",
				Path:        "/tmp/test/abc",
			}
		})

		AfterEach(func() {
			if dirID != "" {
				dirs.Delete(dirID)
			}
		})

		It("Should fail if directory doesn't include the project name", func() {
			var err error
			dirRequest.Path = "/tmp/test2/abc"
			dirID, err := api.GetDirectory(dirRequest)
			Expect(err).To(Equal(app.ErrInvalid))
			Expect(dirID).To(Equal(""))
		})

		It("Should retrieve an existing directory", func() {
			dirRequest.Path = "/tmp/test"
			dirid, err := api.GetDirectory(dirRequest)
			Expect(err).To(BeNil())
			Expect(dirid).To(Equal("test"))
		})

		It("Should create a new directory when it doesn't exist", func() {
			var err error
			dirID, err = api.GetDirectory(dirRequest)
			Expect(err).To(BeNil())
			Expect(dirID).To(ContainSubstring("-"))
		})
	})

	Describe("CreateProject", func() {
		var (
			//projects dai.Projects = dai.NewRProjects(testdb.RSessionMust())
			dirs dai.Dirs = dai.NewRDirs(testdb.RSessionMust())
		)

		BeforeEach(func() {

		})

		AfterEach(func() {

		})

		It("Should fail if project exists and must not exist flag is true", func() {
			req := CreateProjectRequest{
				Name:         "test",
				MustNotExist: true,
			}
			resp, err := api.CreateProject(req)
			Expect(err).To(Equal(app.ErrExists))
			Expect(resp).To(BeNil())
		})

		It("Should succeed if project exists and must not exist flag is false", func() {
			req := CreateProjectRequest{
				Name: "test",
			}
			resp, err := api.CreateProject(req)
			Expect(err).To(BeNil())
			Expect(resp.ProjectID).To(Equal("test"))
			Expect(resp.Existing).To(BeTrue())
		})

		It("Should create a project if it doesn't exist", func() {
			req := CreateProjectRequest{
				Name:         "new-project",
				MustNotExist: true,
			}
			resp, err := api.CreateProject(req)
			Expect(err).To(BeNil())
			Expect(resp.ProjectID).To(ContainSubstring("-"))
			Expect(resp.Existing).To(BeFalse())

			// Delete created project from system.
			session := testdb.RSessionMust()
			var p2d schema.Project2DataDir

			rql := model.ProjectDirs.T().GetAllByIndex("project_id", resp.ProjectID)
			model.ProjectDirs.Qs(session).Row(rql, &p2d)
			dirs.Delete(p2d.DataDirID)
			model.Projects.T().Get(p2d.ProjectID).Delete().RunWrite(session)
		})
	})
})
