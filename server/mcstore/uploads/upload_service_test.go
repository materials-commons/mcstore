package uploads

import (
	"time"

	dmocks "github.com/materials-commons/mcstore/pkg/db/dai/mocks"

	"os"

	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadService", func() {
	var (
		s              *uploadService
		mdirs          *dmocks.Dirs
		muploads       *dmocks.Uploads
		mfiles         *dmocks.Files
		mfiles2        *dmocks.Files2
		req            *UploadRequest
		f              *flow.Request
		savedMCDIRPath string
		fileUpload     schema.FileUpload
		upload         schema.Upload
		nilUpload      *schema.Upload
		nilFile        *schema.File
	)

	BeforeEach(func() {
		savedMCDIRPath = app.MCDir.Path()
		config.Set("MCDIR", "/tmp")
		mdirs = dmocks.NewMDirs()
		muploads = dmocks.NewMUploads()
		mfiles = dmocks.NewMFiles()
		mfiles2 = dmocks.NewMFiles2()
		s = &uploadService{
			files:       mfiles,
			dirs:        mdirs,
			uploads:     muploads,
			tracker:     requestBlockTracker,
			writer:      &blockRequestWriter{},
			requestPath: &mcdirRequestPath{},
			fops:        file.OS,
		}

		f = &flow.Request{
			FlowChunkSize:   1024,
			FlowTotalSize:   int64(len("hello")),
			FlowChunkNumber: 1,
			Chunk:           []byte("hello"),
			FlowIdentifier:  "req",
		}

		req = &UploadRequest{f}

		now := time.Now()

		fileUpload = schema.FileUpload{
			Name:        "uploadtest.txt",
			Checksum:    "checksum",
			Size:        10,
			Birthtime:   now,
			MTime:       now,
			RemoteMTime: now,
			ChunkSize:   10,
			ChunkCount:  1,
		}

		upload = schema.Upload{
			ID:            "req",
			Owner:         "test@mc.org",
			DirectoryID:   "test",
			DirectoryName: "test",
			ProjectID:     "test",
			ProjectOwner:  "test@mc.org",
			ProjectName:   "test",
			Birthtime:     now,
			Host:          "localhost",
			File:          fileUpload,
		}
	})

	AfterEach(func() {
		config.Set("MCDIR", savedMCDIRPath)
	})

	Describe("assemble method tests", func() {
		It("Should return an error when a bad upload id is passed in", func() {
			muploads.On("ByID", "no-such-req").Return(nilUpload, app.ErrNotFound)
			req.FlowIdentifier = "no-such-req"
			uploadFile, err := s.assemble(req, "")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(uploadFile).To(BeNil())
		})

		It("Should return an error and file when it cannot create the destination directory", func() {
			muploads.On("ByID", "req").Return(&upload, nil)
			ifile := &schema.File{
				ID: "uploadtest.txt",
			}
			mfiles2.On("Insert").SetError(nil).SetFile(ifile)
			s.files = mfiles2
			fops := file.MockOps()
			fops.On("MkdirAll").SetError(os.ErrPermission)
			s.fops = fops
			uploadFile, err := s.assemble(req, "dir")
			Expect(err).To(Equal(os.ErrPermission))
			Expect(uploadFile).NotTo(BeNil())
		})

		It("Should return an error and file when the finisher fails", func() {
			muploads.On("ByID", "req").Return(&upload, nil)
			ifile := &schema.File{
				ID: "uploadtest.txt",
			}
			mfiles2.On("Insert").SetError(nil).SetFile(ifile)
			s.files = mfiles2
			fops := file.MockOps()
			fops.On("MkdirAll").SetError(nil)
			s.fops = fops
			s.tracker.load("req", 1)
			s.tracker.addToHash("req", []byte("hello"))
			mfiles2.On("ByPath").SetError(app.ErrNotFound).SetFile(nilFile)
			uploadFile, err := s.assemble(req, "dir")
			Expect(err).NotTo(BeNil())
			Expect(uploadFile).NotTo(BeNil())
		})
	})

	Describe("Upload method tests", func() {
		Context("Successful upload cases", func() {
			// No need to test. The assemble test takes care of
			// most of the steps. And the finish tests will cover
			// the finisher working correctly.
		})
	})
})
