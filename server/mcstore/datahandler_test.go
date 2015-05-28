package mcstore

import (
	"net/http"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"

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

		It("Should return original with converted image when asking for original", func() {
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
})

//func TestServeData(t *testing.T) {
//	mcdir := config.GetString("MCDIR")
//	defer func() {
//		// reset MCDIR to original value when this test ends.
//		config.Set("MCDIR", mcdir)
//	}()
//
//	// Set MCDIR so we know what to test against.
//	config.Set("MCDIR", "/tmp/mcdir")
//
//	a := mocks.NewMAccess()
//	dh := NewDataHandler(a)
//	ts := httptest.NewServer(dh)
//	defer ts.Close()
//
//	// Create response and request
//	req, _ := http.NewRequest("GET", ts.URL, nil)
//	rr := httptest.NewRecorder() // rr = response recorder
//
//	//
//	// Test with no apikey specified
//	//
//	dhhandler := dh.(*dataHandler)
//	path, mediatype, err := dhhandler.serveData(rr, req)
//	require.Equal(t, err, app.ErrNoAccess, "Expected ErrNoAccess: %s ", err)
//	require.Equal(t, path, "", "Got unexpected value for path %s", path)
//	require.Equal(t, mediatype, "", "Got unexpected value for mediatype %s", mediatype)
//
//	fileURL := ts.URL + "/abc-defg-456"
//
//	//
//	// Test with GetFile failing
//	//
//	req, _ = http.NewRequest("GET", fileURL+"?apikey=abc123", nil)
//	var nilFile *schema.File = nil
//	a.On("GetFile", "abc123", "abc-defg-456").Return(nilFile, app.ErrNoAccess)
//	path, mediatype, err = dhhandler.serveData(rr, req)
//	require.Equal(t, err, app.ErrNoAccess, "Expected ErrNoAccess: %s", err)
//	require.Equal(t, path, "", "Got unexpected value for path %s", path)
//	require.Equal(t, mediatype, "", "Got unexpected value for mediatype %s", mediatype)
//
//	//
//	// Test with good key and fileID, get converted image
//	//
//	req, _ = http.NewRequest("GET", fileURL+"?apikey=abc123", nil)
//	f := schema.File{
//		ID: "abc-defg-456",
//		MediaType: schema.MediaType{
//			Mime: "image/tiff",
//		},
//	}
//	a.On("GetFile", "abc123", "abc-defg-456").Return(&f, nil)
//	path, mediatype, err = dhhandler.serveData(rr, req)
//	require.Nil(t, err, "Error should have been nil: %s", err)
//	require.Equal(t, mediatype, "image/jpeg", "Expected image/jpeg, got %s", mediatype)
//	require.Equal(t, path, app.MCDir.FilePathImageConversion(f.FileID()), "Got unexpected value for path %s", path)
//
//	//
//	// Test with good key and fileID, get original image
//	//
//	req, _ = http.NewRequest("GET", fileURL+"?apikey=abc123&original=true", nil)
//	path, mediatype, err = dhhandler.serveData(rr, req)
//	require.Nil(t, err, "Error should have been nil: %s", err)
//	require.Equal(t, mediatype, "image/tiff", "Expected image/tiff, got %s", mediatype)
//	require.Equal(t, path, app.MCDir.FilePath(f.FileID()), "Got unexpected value for path %s", path)
//}
//
//func TestServeHTTP(t *testing.T) {
//	a := mocks.NewMAccess()
//	dh := NewDataHandler(a)
//	ts := httptest.NewServer(dh)
//	defer ts.Close()
//
//	// Create response and request
//	req, _ := http.NewRequest("GET", ts.URL, nil)
//	rr := httptest.NewRecorder() // rr = response recorder
//
//	// Test with no apikey specified
//	dh.ServeHTTP(rr, req)
//	require.Equal(t, rr.Code, http.StatusUnauthorized, "Expected StatusUnauthorized, got %d", rr.Code)
//}
