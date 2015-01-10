package service

import (
	"net/http"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/materials-commons/mcstore/pkg/ws"
)

// dataHandler implements the http.Handler interface. It provides an interface
// to serving up data stored in materials commons.
type dataHandler struct {
	access *domain.Access
}

// NewDataHandler creates a new instance of a dataHandler.
func NewDataHandler(access *domain.Access) http.Handler {
	return &dataHandler{
		access: access,
	}
}

// ServeHTTP serves data stored in materials commons.
func (h *dataHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if err := h.serveData(writer, req); err != nil {
		ws.WriteError(err, writer)
	}
}

// serveData does the actual work of serving the data. It checks the access on each
// request. It will also serve up converted images unless the original flag has been
// specified. The assumption is that without the original flag we are actually trying
// to render an image in a browser. Since browsers do not render all image types we
// convert some types to jpg files. This routine will serve up these jpg conversions
// rather than the original file unless the original flag is specified.
func (h *dataHandler) serveData(writer http.ResponseWriter, req *http.Request) error {
	// All requests require an apikey.
	apikey := req.FormValue("apikey")
	if apikey == "" {
		return app.ErrNoAccess
	}

	// Is the original data requested, or can we serve the converted
	// image data (if it exists)?
	original := getOriginalFormValue(req)
	fileID := filepath.Base(req.URL.Path)

	// Get the file checking its access.
	file, err := h.access.GetFile(apikey, fileID)
	if err != nil {
		return err
	}

	path := filePath(file, original)
	app.Log.Debug(app.Logf("Serving path: %s\n", path))

	writer.Header().Set("Content-Type", file.MediaType.Mime)
	http.ServeFile(writer, req, path)
	return nil
}

// getOriginalFormValue looks for the original argument on the URL. If it exists
// (has a value specified) then return true. The original flag specifies whether
// the user wants the original rather than the converted data.
func getOriginalFormValue(req *http.Request) bool {
	if req.FormValue("original") == "" {
		return false
	}
	return true
}

// filePath determines which file to serve. If there is a converted image we
// will serve that one up unless the original flag is set to true.
func filePath(file *schema.File, original bool) string {
	switch {
	case isConvertedImage(file.MediaType.Mime) && original:
		return imageConversionPath(file.FileID())
	default:
		return app.MCDir.FilePath(file.FileID())
	}
}

// isTiff checks a name to see if it is for a TIFF file.
func isConvertedImage(mime string) bool {
	switch mime {
	case "image/tiff":
		return true
	case "image/x-ms-bmp":
		return true
	default:
		return false
	}
}

// imageConversionPath returns the path to the converted image. Converted images
// are kept in the filepath subdirectory .conversion.
func imageConversionPath(id string) string {
	return filepath.Join(app.MCDir.FileDir(id), ".conversion", id+".jpg")
}
