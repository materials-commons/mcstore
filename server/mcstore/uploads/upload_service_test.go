package uploads

import (
	dmocks "github.com/materials-commons/mcstore/pkg/db/dai/mocks"

	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadService", func() {
	var (
		s              *uploadService
		mdirs          *dmocks.Dirs
		muploads       *dmocks.Uploads
		mfiles         *dmocks.Files
		req            *UploadRequest
		f              *flow.Request
		savedMCDIRPath string
	)

	BeforeEach(func() {
		savedMCDIRPath = app.MCDir.Path()
		config.Set("MCDIR", "/tmp")
		mdirs = dmocks.NewMDirs()
		muploads = dmocks.NewMUploads()
		mfiles = dmocks.NewMFiles()
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
	})

	AfterEach(func() {
		config.Set("MCDIR", savedMCDIRPath)
	})

	Describe("Upload method tests", func() {
		Context("Successful upload cases", func() {
			It("Should succeed for a single block upload", func() {
				str := ""
				Expect(str).To(Equal(""))
			})
		})

		Context("Failed upload cases", func() {
			It("Should succeed for a multiple block upload", func() {

			})
		})
	})

	Describe("assemble method tests", func() {

	})

})
