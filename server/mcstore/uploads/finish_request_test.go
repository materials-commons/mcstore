package uploads

import (
	"github.com/materials-commons/mcstore/pkg/app"
	dmocks "github.com/materials-commons/mcstore/pkg/db/dai/mocks"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FinishRequest", func() {
	var (
		files *dmocks.Files
		dirs  *dmocks.Dirs
		f     *finisher
	)

	BeforeEach(func() {
		files = dmocks.NewMFiles()
		dirs = dmocks.NewMDirs()
		f = &finisher{
			files: files,
			dirs:  dirs,
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
			var nilFile *schema.File = nil
			files.On("ByPath", "file.name", "dir").Return(nilFile, app.ErrNotFound)
			parentID, err := f.parentID("file.name", "dir")
			Expect(err).To(BeNil())
			Expect(parentID).To(Equal(""))
		})

		It("Should return an error when ByPath returns an error other than app.ErrNotFound", func() {
			var nilFile *schema.File = nil
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
	})
})
