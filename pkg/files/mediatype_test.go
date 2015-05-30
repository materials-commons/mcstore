package files

import (
	"mime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"path/filepath"

	"fmt"

	"github.com/rakyll/magicmime"
)

var (
	testDataDir          string
	bmpFileNoExtension   string
	tiffFileNoExtension  string
	jpgFileWithExtension string
)

func init() {
	testDataDir, _ = filepath.Abs("../../test-data")
	bmpFileNoExtension = filepath.Join(testDataDir, "bm", "p1", "xxxx-bmp123")
	tiffFileNoExtension = filepath.Join(testDataDir, "ti", "f1", "xxxx-tif123")
	jpgFileWithExtension = filepath.Join(testDataDir, "xxxx-tif123.jpg")
}

var _ = Describe("MediaType", func() {
	Describe("Test underlying dependencies", func() {
		Context("mime.TypeByExtension", func() {
			It("Should return empty string for unknown types", func() {
				mtype := mime.TypeByExtension(".rhit")
				Expect(mtype).To(Equal(""))
			})

			It("Should return a mime type for .jpg extension", func() {
				mtype := mime.TypeByExtension(".jpg")
				Expect(mtype).To(Equal("image/jpeg"))
			})
		})

		Context("magicmime TypeByFile", func() {
			It("Should detect type of file without an extension", func() {
				magic, err := magicmime.New(magicmime.MAGIC_MIME)
				Expect(err).To(BeNil())
				fmt.Println("testDataDir", testDataDir)
				ftype, err := magic.TypeByFile(bmpFileNoExtension)
				Expect(err).To(BeNil())
				Expect(ftype).To(Equal("image/x-ms-bmp; charset=binary"))
			})
		})
	})

	Describe("MediaType Method Tests", func() {
		It("Should detect the BMP file type", func() {
			mediatype := MediaType("xxxx-bmp123", bmpFileNoExtension)
			Expect(mediatype.Mime).To(Equal("image/x-ms-bmp"))
			Expect(mediatype.Description).To(Equal("BMP"))
		})

		It("Should detect the TIFF file type", func() {
			mediatype := MediaType("xxxx-tif123", tiffFileNoExtension)
			Expect(mediatype.Mime).To(Equal("image/tiff"))
			Expect(mediatype.Description).To(Equal("TIFF"))
		})

		It("Should detect a JPEG file type", func() {
			mediatype := MediaType("xxxx-tif123.jpg", jpgFileWithExtension)
			Expect(mediatype.Mime).To(Equal("image/jpeg"))
			Expect(mediatype.Description).To(Equal("JPEG"))
		})

		It("Should return unknown on a bad file path with no extension", func() {
			mediatype := MediaType("xxxx-bmp123", "/does/not/exist/xxxx-bmp123")
			Expect(mediatype.Mime).To(Equal("unknown"))
			Expect(mediatype.Description).To(Equal("Unknown"))
		})

		It("Should detect matlab files by their extension", func() {
			mediatype := MediaType("abc.m", "/tmp/abc.m")
			Expect(mediatype.Mime).To(Equal("application/matlab"))
			Expect(mediatype.Description).To(Equal("Matlab"))
		})
	})
})
