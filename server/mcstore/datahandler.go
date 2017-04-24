package mcstore

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
	access domain.Access
}

// NewDataHandler creates a new instance of a dataHandler.
func NewDataHandler(access domain.Access) http.Handler {
	return &dataHandler{
		access: access,
	}
}

// ServeHTTP serves data stored in materials commons.
func (h *dataHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	path, mediaType, err := h.serveData(writer, req)
	switch {
	case err != nil:
		ws.WriteError(err, writer)
	default:
		serveFile(writer, req, path, mediaType)
	}
}

// serveFile serves up the actual file contents. It sets the content-type header to
// the mediatype specified. If the file doesn't exist, or the server doesn't have
// permissions to access the file then a 404 (not found) will be returned by the
// http.ServeFile method.
func serveFile(writer http.ResponseWriter, req *http.Request, path, mediatype string) {
	writer.Header().Add("Content-Type", mediatype)
	if mediatype == "application/pdf" {
		writer.Header().Add("Content-Disposition", `inline; filename="filename.pdf"`)
	}
	app.Log.Debugf("Set Content-Type to %s", mediatype)
	http.ServeFile(writer, req, path)
}

// serveData does the actual work of serving the data. It checks the access on each
// request. It will also serve up converted images unless the original flag has been
// specified. The assumption is that without the original flag we are actually trying
// to render an image in a browser. Since browsers do not render all image types we
// convert some types to jpg files. This routine will serve up these jpg conversions
// rather than the original file unless the original flag is specified.
func (h *dataHandler) serveData(writer http.ResponseWriter, req *http.Request) (path string, mediatype string, err error) {
	// All requests require an apikey.
	apikey := req.FormValue("apikey")
	if apikey == "" {
		return path, mediatype, app.ErrNoAccess
	}
	app.Log.Debugf("serveData - Request for apikey %s", apikey)

	// Is the original data requested, or can we serve the converted
	// image data (if it exists)?
	original := getOriginalFormValue(req)
	app.Log.Debugf("serveData - Original flag %t", original)

	fileID := filepath.Base(req.URL.Path)
	app.Log.Debugf("serveData - fileID %s, URL %s", fileID, req.URL.Path)

	// Get the file, checking its access.
	file, err := h.access.GetFile(apikey, fileID)
	if err != nil {
		return path, mediatype, err
	}

	path = filePath(file, original)
	app.Log.Debugf("serveData - Serving path: %s\n", path)

	mediatype = file.MediaType.Mime

	// The content type is dependent on whether we are
	// serving the original or the converted file.
	if !original && isConvertedImage(file.MediaType.Mime) {
		mediatype = "image/jpeg"
	} else if !original && isExcelSpreadsheet(file.MediaType.Mime) {
		mediatype = "application/pdf"
	}

	return path, mediatype, nil
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
	case isConvertedImage(file.MediaType.Mime) && !original:
		return app.MCDir.FilePathImageConversion(file.FileID())
	case isExcelSpreadsheet(file.MediaType.Mime) && !original:
		return app.MCDir.FilePathFromConversionToPDF(file.FileID())
	default:
		return app.MCDir.FilePath(file.FileID())
	}
}

// isConvertedImage checks a name to see if it is an image type we have converted.
func isConvertedImage(mime string) bool {
	switch mime {
	case "image/tiff":
		return true
	case "image/x-ms-bmp":
		return true
	case "image/bmp":
		return true
	default:
		return false
	}
}

// isExcelSpreadsheet checks the mime type to see if it is an excel spreadsheet
func isExcelSpreadsheet(mime string) bool {
	switch mime {
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return true
	case "application/vnd.MS-Excel":
		return true
	default:
		return false
	}
}
