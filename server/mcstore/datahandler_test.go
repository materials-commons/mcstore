package mcstore

import (
	"net/http"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"

	"net/http/httptest"

	"github.com/materials-commons/mcstore/pkg/domain/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DataHandler", func() {
	Describe("getOriginalFormValue Method Tests", func() {
		It("Should return false if original flag is not given", func() {
			req, _ := http.NewRequest("GET", "http://localhost", nil)
			Expect(getOriginalFormValue(req)).To(BeFalse())
		})

		It("Should return false if original flag is passed with no value", func() {
			req, _ := http.NewRequest("GET", "http://localhost?original", nil)
			Expect(getOriginalFormValue(req)).To(BeFalse())
		})

		Context("Setting original flag to any value should always return true", func() {
			It("Should return true for numeric true and false (0, 1)", func() {
				req, _ := http.NewRequest("GET", "http://localhost?original=0", nil)
				Expect(getOriginalFormValue(req)).To(BeTrue())

				req, _ = http.NewRequest("GET", "http://localhost?original=1", nil)
				Expect(getOriginalFormValue(req)).To(BeTrue())
			})

			It("Should return true for boolean true and false", func() {
				req, _ := http.NewRequest("GET", "http://localhost?original=false", nil)
				Expect(getOriginalFormValue(req)).To(BeTrue())

				req, _ = http.NewRequest("GET", "http://localhost?original=true", nil)
				Expect(getOriginalFormValue(req)).To(BeTrue())
			})
		})

		It("Should work if original is passed in as the second flag", func() {
			req, _ := http.NewRequest("GET", "http://localhost?apikey=abc&original=true", nil)
			Expect(getOriginalFormValue(req)).To(BeTrue())
		})
	})

	Describe("isConvertedImage Method Tests", func() {
		It("Should return true for image/tiff", func() {
			Expect(isConvertedImage("image/tiff")).To(BeTrue())
		})

		It("Should return true for image/x-ms-bmp", func() {
			Expect(isConvertedImage("image/x-ms-bmp")).To(BeTrue())
		})

		It("Should return true for image/bmp", func() {
			Expect(isConvertedImage("image/bmp")).To(BeTrue())
		})

		It("Should return false for image/jpg", func() {
			Expect(isConvertedImage("image/jpg")).To(BeFalse())
		})
	})

	Describe("filePath Method Tests", func() {
		var (
			saved string
			f     schema.File = schema.File{
				ID: "abc-defg-456",
				MediaType: schema.MediaType{
					Mime: "image/tiff",
				},
			}
		)

		BeforeEach(func() {
			saved = config.GetString("MCDIR")
			config.Set("MCDIR", "/tmp/mcdir")
		})

		AfterEach(func() {
			config.Set("MCDIR", saved)
		})

		It("Should return converted and not original for tiff images", func() {
			f.MediaType.Mime = "image/tiff"
			path := filePath(&f, false)
			Expect(path).To(Equal(app.MCDir.FilePathImageConversion(f.FileID())))
		})

		It("Should return original even though converted is available when asking for original", func() {
			f.MediaType.Mime = "image/tiff"
			path := filePath(&f, true)
			Expect(path).To(Equal(app.MCDir.FilePath(f.FileID())))
		})

		It("Should return original when unconverted type and not requesting original", func() {
			f.MediaType.Mime = "text/plain"
			path := filePath(&f, false)
			Expect(path).To(Equal(app.MCDir.FilePath(f.FileID())))
		})

		It("Should return original when unconverted and requesting original", func() {
			f.MediaType.Mime = "text/plain"
			path := filePath(&f, true)
			Expect(path).To(Equal(app.MCDir.FilePath(f.FileID())))
		})

		It("Should return usesID converted path when uses is set and not requesting original on tiff", func() {
			f.MediaType.Mime = "image/tiff"
			f.UsesID = "def-ghij-789"
			path := filePath(&f, false)
			Expect(path).To(Equal(app.MCDir.FilePathImageConversion(f.FileID())))
		})

		It("Should return usesID original path when uses is set and requesting original on tiff", func() {
			f.MediaType.Mime = "image/tiff"
			f.UsesID = "def-ghij-789"
			path := filePath(&f, true)
			Expect(path).To(Equal(app.MCDir.FilePath(f.FileID())))
		})

		It("Should return original with uses set not requesting original for text/plain (non-converted)", func() {
			f.MediaType.Mime = "text/plain"
			f.UsesID = "def-ghij-789"
			path := filePath(&f, false)
			Expect(path).To(Equal(app.MCDir.FilePath(f.FileID())))
		})

		It("Should return original with uses set, requesting original for non-converted type", func() {
			f.MediaType.Mime = "text/plain"
			f.UsesID = "def-ghij-789"
			path := filePath(&f, true)
			Expect(path).To(Equal(app.MCDir.FilePath(f.FileID())))
		})

	})

	Describe("serveData Method Tests", func() {
		var (
			server      *httptest.Server
			saved       string = config.GetString("MCDIR")
			rr          *httptest.ResponseRecorder
			datahandler http.Handler
			access      *mocks.Access
			dhhandler   *dataHandler
		)

		BeforeEach(func() {
			access = mocks.NewMAccess()
			datahandler = NewDataHandler(access)
			dhhandler = datahandler.(*dataHandler)
			server = httptest.NewServer(datahandler)
			rr = httptest.NewRecorder()
			config.Set("MCDIR", "/tmp/mcdir")
		})

		AfterEach(func() {
			server.Close()
			config.Set("MCDIR", saved)
		})

		It("Should fail if no apikey is specified", func() {
			req, _ := http.NewRequest("GET", server.URL, nil)
			path, mediatype, err := dhhandler.serveData(rr, req)
			Expect(err).To(Equal(app.ErrNoAccess), "Expected ErrNoAccess, got: %s ", err)
			Expect(path).To(Equal(""), "Got unexpected value for path %s", path)
			Expect(mediatype).To(Equal(""), "Got unexpected value for mediatype %s", mediatype)
		})

		It("Should fail when user doesn't have access to the requested file", func() {
			fileURL := server.URL + "/abc-defg-456?apikey=abc123"
			req, _ := http.NewRequest("GET", fileURL, nil)
			var nilFile *schema.File
			access.On("GetFile", "abc123", "abc-defg-456").Return(nilFile, app.ErrNoAccess)
			path, mediatype, err := dhhandler.serveData(rr, req)
			Expect(err).To(Equal(app.ErrNoAccess), "Expected ErrNoAccess: %s", err)
			Expect(path).To(Equal(""), "Got unexpected value for path %s", path)
			Expect(mediatype).To(Equal(""), "Got unexpected value for mediatype %s", mediatype)
		})

		It("Should succeed with a good key and fileID and return converted image for a tiff", func() {
			fileURL := server.URL + "/abc-defg-456?apikey=abc123"
			req, _ := http.NewRequest("GET", fileURL, nil)
			f := schema.File{
				ID: "abc-defg-456",
				MediaType: schema.MediaType{
					Mime: "image/tiff",
				},
			}

			access.On("GetFile", "abc123", "abc-defg-456").Return(&f, nil)
			path, mediatype, err := dhhandler.serveData(rr, req)
			Expect(err).To(BeNil())
			Expect(mediatype).To(Equal("image/jpeg"), "Expected image/jpeg, got %s", mediatype)
			Expect(path).To(Equal(app.MCDir.FilePathImageConversion(f.FileID())), "Got unexpected value for path %s", path)
		})

		It("Should succeed with a good key and fileID, and return the original image when original flag set for tiff", func() {
			fileURL := server.URL + "/abc-defg-456?apikey=abc123&original=true"
			req, _ := http.NewRequest("GET", fileURL, nil)
			f := schema.File{
				ID: "abc-defg-456",
				MediaType: schema.MediaType{
					Mime: "image/tiff",
				},
			}

			access.On("GetFile", "abc123", "abc-defg-456").Return(&f, nil)
			path, mediatype, err := dhhandler.serveData(rr, req)
			Expect(err).To(BeNil())
			Expect(mediatype).To(Equal("image/tiff"), "Expected image/tiff, got %s", mediatype)
			Expect(path).To(Equal(app.MCDir.FilePath(f.FileID())), "Got unexpected value for path %s", path)
		})
	})
})
