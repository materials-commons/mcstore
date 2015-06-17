package uploads

import (
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	dmocks "github.com/materials-commons/mcstore/pkg/db/dai/mocks"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/testdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FinishRequest", func() {
	var (
		mfiles  *dmocks.Files
		mdirs   *dmocks.Dirs
		mfiles2 *dmocks.Files2
		fops    *file.MockOperations
		f       *finisher
		nilFile *schema.File = nil
		files   dai.Files
		dirs    dai.Dirs
	)

	BeforeEach(func() {
		mfiles = dmocks.NewMFiles()
		mdirs = dmocks.NewMDirs()
		mfiles2 = dmocks.NewMFiles2()
		fops = file.MockOps()
		f = &finisher{
			files: mfiles,
			dirs:  mdirs,
			fops:  fops,
		}
	})

	Describe("parentID method tests", func() {
		It("Should return a parent when there is one", func() {
			parentFile := &schema.File{
				ID: "parent",
			}
			mfiles.On("ByPath", "file.name", "dir").Return(parentFile, nil)
			parentID, err := f.parentID("file.name", "dir")
			Expect(err).To(BeNil())
			Expect(parentID).To(Equal("parent"))
		})

		It("Should not return a parent when there isn't one", func() {
			mfiles.On("ByPath", "file.name", "dir").Return(nilFile, app.ErrNotFound)
			parentID, err := f.parentID("file.name", "dir")
			Expect(err).To(BeNil())
			Expect(parentID).To(Equal(""))
		})

		It("Should return an error when ByPath returns an error other than app.ErrNotFound", func() {
			mfiles.On("ByPath", "file.name", "dir").Return(nilFile, app.ErrInvalid)
			parentID, err := f.parentID("file.name", "dir")
			Expect(err).To(Equal(app.ErrInvalid))
			Expect(parentID).To(Equal(""))
		})
	})

	Describe("fileInDir method tests", func() {
		It("Should return false if the file isn't in the directory", func() {
			var noFiles []schema.File
			mdirs.On("Files", "dir").Return(noFiles, app.ErrNotFound)
			Expect(f.fileInDir("checksum", "file.name", "dir")).To(BeFalse())
		})

		It("Should return false if there is a matching file with a different checksum", func() {
			matchingFile := schema.File{
				Name:     "file.name",
				Checksum: "wrongchecksum",
			}
			var matching []schema.File = []schema.File{matchingFile}
			mdirs.On("Files", "dir").Return(matching, nil)
			Expect(f.fileInDir("abc123", "file.name", "dir")).To(BeFalse())
		})

		It("Should return true if the file with exact checksum is in the directory", func() {
			matchingFile := schema.File{
				Name:     "file.name",
				Checksum: "abc123",
			}
			var matching []schema.File = []schema.File{matchingFile}
			mdirs.On("Files", "dir").Return(matching, nil)
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

		Context("Simulate connection to database", func() {
			It("Should fail if the size is wrong.", func() {
				mfiles.On("ByPath", "file.name", "dir").Return(nilFile, app.ErrNotFound)
				mFileInfo := file.MockFileInfo{
					MSize: 2,
				}
				freq.FlowTotalSize = 3
				fops.On("Stat").SetError(nil).SetValue(mFileInfo)
				err := f.finish(req, "fileID", "checksum", upload)
				Expect(err).To(Equal(app.ErrInvalid))
			})

		})

		Context("Connect to database", func() {
			BeforeEach(func() {
				files = dai.NewRFiles(testdb.RSession())
				dirs = dai.NewRDirs(testdb.RSession())

				// insert file we are going to test against
				tfile := schema.NewFile("testfile1.txt", "test@mc.org")
				tfile.ID = "testfile1.txt"
				tfile.Size = 100
				files.Insert(&tfile, "test", "test")
			})

			AfterEach(func() {
				// delete file that was used for testing
				files.Delete("testfile1.txt", "test", "test")
			})

			It("Should succeed if no matching checksum is found", func() {
				mFileInfo := file.MockFileInfo{
					MSize: 100,
				}
				fops.On("Stat").SetError(nil).SetValue(mFileInfo)
				f.files = files
				f.dirs = dirs
				req.FlowTotalSize = 100
				err := f.finish(req, "testfile1.txt", "no-matching-checksum", upload)
				Expect(err).To(BeNil())
				updatedFile, err := files.ByID("testfile1.txt")
				Expect(err).To(BeNil())
				Expect(updatedFile.Size).To(BeNumerically("==", 100))
				Expect(updatedFile.Checksum).To(Equal("no-matching-checksum"))
				Expect(updatedFile.Uploaded).To(BeNumerically("==", 100))
				Expect(updatedFile.Current).To(BeTrue())
			})

			It("Should succeed if matching checksum is found", func() {
				mFileInfo := file.MockFileInfo{
					MSize: 100,
				}
				fops.On("Stat").SetError(nil).SetValue(mFileInfo)
				f.files = files
				f.dirs = dirs
				req.FlowTotalSize = 100
				err := f.finish(req, "testfile1.txt", "no-matching-checksum", upload)
				Expect(err).To(BeNil())
				updatedFile, err := files.ByID("testfile1.txt")
				Expect(err).To(BeNil())
				Expect(updatedFile.Size).To(BeNumerically("==", 100))
				Expect(updatedFile.Checksum).To(Equal("no-matching-checksum"))
				Expect(updatedFile.Uploaded).To(BeNumerically("==", 100))
				Expect(updatedFile.Current).To(BeTrue())
			})
		})
	})
})
