package service

import (
	"net/http"

	"github.com/materials-commons/mcstore/pkg/app"
)

type dataHandler struct {
}

func NewDataHandler() http.Handler {
	return &dataHandler{}
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

	_ = req.FormValue("download")
	return nil
}
