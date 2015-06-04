package uploads

import (
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	dmocks "github.com/materials-commons/mcstore/pkg/db/dai/mocks"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FinishRequest", func() {
	var (
		files   *dmocks.Files
		dirs    *dmocks.Dirs
		files2  *dmocks.Files2
		fops    *file.MockOperations
		f       *finisher
		nilFile *schema.File = nil
	)

	BeforeEach(func() {
		files = dmocks.NewMFiles()
		dirs = dmocks.NewMDirs()
		files2 = dmocks.NewMFiles2()
		fops = file.MockOps()
		f = &finisher{
			files: files,
			dirs:  dirs,
			fops:  fops,
		}
	})

	Describe("parentID method tests", func() {
		It("Should return a parent when there is one", func() {
			parentFile := &schema.File{
				ID: "parent",
			}
			files.On("ByPath", "file.name", "dir").Return(parentFile, nil)
			parentID, err := f.parentID("file.name", "dir")
			Expect(err).To(BeNil())
			Expect(parentID).To(Equal("parent"))
		})

		It("Should not return a parent when there isn't one", func() {
			files.On("ByPath", "file.name", "dir").Return(nilFile, app.ErrNotFound)
			parentID, err := f.parentID("file.name", "dir")
			Expect(err).To(BeNil())
			Expect(parentID).To(Equal(""))
		})

		It("Should return an error when ByPath returns an error other than app.ErrNotFound", func() {
			files.On("ByPath", "file.name", "dir").Return(nilFile, app.ErrInvalid)
			parentID, err := f.parentID("file.name", "dir")
			Expect(err).To(Equal(app.ErrInvalid))
			Expect(parentID).To(Equal(""))
		})
	})

	Describe("fileInDir method tests", func() {
		It("Should return false if the file isn't in the directory", func() {
			var noFiles []schema.File
			dirs.On("Files", "dir").Return(noFiles, app.ErrNotFound)
			Expect(f.fileInDir("checksum", "file.name", "dir")).To(BeFalse())
		})

		It("Should return false if there is a matching file with a different checksum", func() {
			matchingFile := schema.File{
				Name:     "file.name",
				Checksum: "wrongchecksum",
			}
			var matching []schema.File = []schema.File{matchingFile}
			dirs.On("Files", "dir").Return(matching, nil)
			Expect(f.fileInDir("abc123", "file.name", "dir")).To(BeFalse())
		})

		It("Should return true if the file with exact checksum is in the directory", func() {
			matchingFile := schema.File{
				Name:     "file.name",
				Checksum: "abc123",
			}
			var matching []schema.File = []schema.File{matchingFile}
			dirs.On("Files", "dir").Return(matching, nil)
			Expect(f.fileInDir("abc123", "file.name", "dir")).To(BeTrue())
		})
	})

	Describe("finish method tests", func() {
		var (
			upload *schema.Upload = &schema.Upload{
				DirectoryID: "dir",
				File: schema.FileUpload{
					Name: "file.name",
				},
			}
			freq *flow.Request
			req  *UploadRequest
		)

		BeforeEach(func() {
			freq = &flow.Request{}
			req = &UploadRequest{freq}
		})

		It("Should fail if the size is wrong.", func() {
			files.On("ByPath", "file.name", "dir").Return(nilFile, app.ErrNotFound)
			mFileInfo := file.MockFileInfo{
				MSize: 2,
			}
			freq.FlowTotalSize = 3
			fops.On("Stat").SetError(nil).SetValue(mFileInfo)
			err := f.finish(req, "fileID", "checksum", upload)
			Expect(err).To(Equal(app.ErrInvalid))
		})
	})
})
