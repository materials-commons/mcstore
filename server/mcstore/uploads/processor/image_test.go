package processor

import (
	"crypto/md5"
	"path/filepath"

	"os"

	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Image", func() {
	Describe("Process Method Tests", func() {
		Context("Create JPEG from am image", func() {
			var (
				savedMCDIRPath string
				testMCDIRPath  string
			)

			const tiffFileID = "xxxx-tif123"
			const bmpFileID = "xxxx-bmp123"

			BeforeEach(func() {
				savedMCDIRPath = app.MCDir.Path()
				testMCDIRPath, _ = filepath.Abs("../../../../test-data")
				config.Set("MCDIR", testMCDIRPath)
			})

			AfterEach(func() {
				os.RemoveAll(filepath.Join(app.MCDir.FileDir(tiffFileID), ".conversion"))
				os.RemoveAll(filepath.Join(app.MCDir.FileDir(bmpFileID), ".conversion"))
				config.Set("MCDIR", savedMCDIRPath)
			})

			It("Should create a valid JPEG image file from a TIFF file", func() {
				imageProcessor := newImageFileProcessor(tiffFileID)
				err := imageProcessor.Process()
				Expect(err).To(BeNil())
				expectedHash, _ := file.HashStr(md5.New(), filepath.Join(testMCDIRPath, tiffFileID+".jpg"))
				generatedHash, _ := file.HashStr(md5.New(), filepath.Join(app.MCDir.FileDir(tiffFileID), ".conversion", tiffFileID+".jpg"))
				Expect(expectedHash).To(Equal(generatedHash))
			})

			It("Should create a valid JPEG from a BMP file", func() {
				imageProcessor := newImageFileProcessor(bmpFileID)
				err := imageProcessor.Process()
				Expect(err).To(BeNil())
				expectedHash, _ := file.HashStr(md5.New(), filepath.Join(testMCDIRPath, bmpFileID+".jpg"))
				generatedHash, _ := file.HashStr(md5.New(), filepath.Join(app.MCDir.FileDir(bmpFileID), ".conversion", bmpFileID+".jpg"))
				Expect(expectedHash).To(Equal(generatedHash))
			})
		})
	})
})
