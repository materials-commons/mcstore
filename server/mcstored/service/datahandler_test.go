package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetOriginal(t *testing.T) {
	// handler := func(w http.ResponseWriter, r *http.Request) {
	// 	// Nothing to do
	// 	http.Error(w, "nothing", http.StatusInternalServerError)
	// }
	//	ts := httptest.
	req, _ := http.NewRequest("GET", "http://localhost", nil)

	// Test with no flag
	require.False(t, getOriginalFormValue(req), "No original flag specified, but returned true")

	// Test with flag set to no value
	req, _ = http.NewRequest("GET", "http://localhost?original", nil)
	require.False(t, getOriginalFormValue(req), "No original flag specified, but returned true")

	// Test with flag set to any value
	req, _ = http.NewRequest("GET", "http://localhost?original='1'", nil)
	require.True(t, getOriginalFormValue(req), "Original flag specified with value, but returned false")

	// Test original as second flag
	req, _ = http.NewRequest("GET", "http://localhost?apikey=abc&original=true", nil)
	require.True(t, getOriginalFormValue(req), "Original flag specified with value, but returned false")
}

func TestImageConversionPath(t *testing.T) {
	mcdir := config.GetString("MCDIR")
	defer func() {
		// reset MCDIR to original value when this test ends.
		config.Set("MCDIR", mcdir)
	}()

	// Set MCDIR so we know what to test against.
	config.Set("MCDIR", "/tmp/mcdir")

	// imageConversionPath should always be called with a good id. So we
	// just need to test that the directory it constructs looks ok.
	fileID := "abc-defg-456"
	path := imageConversionPath(fileID)
	require.Equal(t, "/tmp/mcdir/de/fg/.conversion/abc-defg-456.jpg", path, "Got unexpected path: %s", path)
}

func TestIsConvertedImage(t *testing.T) {
	// Test against a couple of different MiME types.
	require.True(t, isConvertedImage("image/tiff"), "image/tiff should be a converted type")
	require.True(t, isConvertedImage("image/x-ms-bmp"), "image/x-ms-bmp should be a converted type")
	require.False(t, isConvertedImage("image/jpg"), "image/jpg should not be converted")
}

func TestFilePath(t *testing.T) {
	mcdir := config.GetString("MCDIR")
	defer func() {
		// reset MCDIR to original value when this test ends.
		config.Set("MCDIR", mcdir)
	}()

	// Set MCDIR so we know what to test against.
	config.Set("MCDIR", "/tmp/mcdir")

	// All we need is a file with a mediatype, the other entries
	// don't matter
	f := schema.File{
		ID: "abc-defg-456",
		MediaType: schema.MediaType{
			Mime: "image/tiff",
		},
	}

	// Test converted image, and not requesting original
	path := filePath(&f, false)
	require.Equal(t, path, "/tmp/mcdir/de/fg/.conversion/abc-defg-456.jpg")

	// Test converted image and requesting original
	path = filePath(&f, true)
	require.Equal(t, path, "/tmp/mcdir/de/fg/abc-defg-456")

	// Test unconverted and not requesting original
	f.MediaType.Mime = "text/plain"
	path = filePath(&f, false)
	require.Equal(t, path, "/tmp/mcdir/de/fg/abc-defg-456")

	// Test unconverted and requesting original
	path = filePath(&f, true)
	require.Equal(t, path, "/tmp/mcdir/de/fg/abc-defg-456")

	// Test with uses set, converted image, not requesting original
	f.MediaType.Mime = "image/tiff"
	f.UsesID = "def-ghij-789"
	path = filePath(&f, false)
	require.Equal(t, path, "/tmp/mcdir/gh/ij/.conversion/def-ghij-789.jpg")

	// Test with uses set, converted image, requesting original
	path = filePath(&f, true)
	require.Equal(t, path, "/tmp/mcdir/gh/ij/def-ghij-789")

	// Test with uses set, not converted image, not requesting original
	f.MediaType.Mime = "text/plain"
	path = filePath(&f, false)
	require.Equal(t, path, "/tmp/mcdir/gh/ij/def-ghij-789")

	// Test with uses set, not converted image, requesting original
	path = filePath(&f, true)
	require.Equal(t, path, "/tmp/mcdir/gh/ij/def-ghij-789")
}

func TestServeData(t *testing.T) {
	a := mocks.NewMAccess()
	dh := NewDataHandler(a)
	ts := httptest.NewServer(dh)
	defer ts.Close()

	// Create response and request
	req, _ := http.NewRequest("GET", ts.URL, nil)
	rr := httptest.NewRecorder() // rr = response recorder

	// Test with no apikey specified
	dhhandler := dh.(*dataHandler)
	_, mediatype, err := dhhandler.serveData(rr, req)
	require.Equal(t, err, app.ErrNoAccess, "Expected ErrNoAccess: %s ", err)

	fileURL := ts.URL + "/abc-defg-456"

	// Test with GetFile failing
	req, _ = http.NewRequest("GET", fileURL+"?apikey=abc123", nil)
	var nilFile *schema.File = nil
	a.On("GetFile", "abc123", "abc-defg-456").Return(nilFile, app.ErrNoAccess)
	_, mediatype, err = dhhandler.serveData(rr, req)
	require.Equal(t, err, app.ErrNoAccess, "Expected ErrNoAccess: %s", err)

	// Test with good key and fileID, get converted image
	req, _ = http.NewRequest("GET", fileURL+"?apikey=abc123", nil)
	f := schema.File{
		ID: "abc-defg-456",
		MediaType: schema.MediaType{
			Mime: "image/tiff",
		},
	}
	a.On("GetFile", "abc123", "abc-defg-456").Return(&f, nil)
	_, mediatype, err = dhhandler.serveData(rr, req)
	require.Nil(t, err, "Error should have been nil: %s", err)
	require.Equal(t, mediatype, "image/jpeg", "Expected image/jpeg, got %s", mediatype)

	// Test with good key and fileID, get original image
	req, _ = http.NewRequest("GET", fileURL+"?apikey=abc123&original=true", nil)
	_, mediatype, err = dhhandler.serveData(rr, req)
	require.Nil(t, err, "Error should have been nil: %s", err)
	require.Equal(t, mediatype, "image/tiff", "Expected image/tiff, got %s", mediatype)
}

func TestServeHTTP(t *testing.T) {
	a := mocks.NewMAccess()
	dh := NewDataHandler(a)
	ts := httptest.NewServer(dh)
	defer ts.Close()

	// Create response and request
	req, _ := http.NewRequest("GET", ts.URL, nil)
	rr := httptest.NewRecorder() // rr = response recorder

	// Test with no apikey specified
	dh.ServeHTTP(rr, req)
	require.Equal(t, rr.Code, http.StatusUnauthorized, "Expected StatusUnauthorized, got %d", rr.Code)
}
