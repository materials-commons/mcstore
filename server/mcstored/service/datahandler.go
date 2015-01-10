package service

import (
	"net/http"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/domain"
)

type dataHandler struct {
	access *domain.Access
}

func NewDataHandler(access *domain.Access) http.Handler {
	return &dataHandler{
		access: access,
	}
}

func (h *dataHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if err := h.serveData(writer, req); err != nil {
		// Write error out
	}
}

func (h *dataHandler) serveData(writer http.ResponseWriter, req *http.Request) error {
	apikey := req.FormValue("apikey")
	if apikey == "" {
		return app.ErrNoAccess
	}

	download := req.FormValue("download")
	fileID := filepath.Base(req.URL.Path)
	file, err := h.access.GetFile(apikey, fileID)
	if err != nil {
		return err
	}

	var path string
	if isConvertedImage(file.MediaType.Mime) && download == "" {
		path = imageConversionPath(file.FileID())
	} else {
		path = app.MCDir.FilePath(file.FileID())
	}
	writer.Header().Set("Content-Type", file.MediaType.Mime)
	app.Log.Debug(app.Logf("Serving path: %s\n", path))
	http.ServeFile(writer, req, path)

	return nil
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

func imageConversionPath(id string) string {
	return filepath.Join(app.MCDir.FileDir(id), ".conversion", id+".jpg")
}
