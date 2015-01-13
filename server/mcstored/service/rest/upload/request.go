package upload

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

type RequestWriter interface {
	Write(req *flow.Request) error
}

type RequestPath interface {
	Path(req *flow.Request) string
	Dir(req *flow.Request) string
}

type fileRequestWriter struct {
	RequestPath
}

func newFileRequestWriter(requestPath RequestPath) *fileRequestWriter {
	return &fileRequestWriter{
		RequestPath: requestPath,
	}
}

func (r *fileRequestWriter) Write(req *flow.Request) error {
	path := r.Path(req)
	err := r.validateWrite(path, req)
	switch {
	case err == nil:
		return ioutil.WriteFile(path, req.Chunk, 0700)
	case err == app.ErrExists:
		return nil
	default:
		return err
	}
}

// validateWrite determines if a particular chunk can be written.
// If the size of the ondisk chunk is smaller than the request
// chunk then that chunk is incomplete and we allow a write to it.
func (r *fileRequestWriter) validateWrite(path string, req *flow.Request) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	finfo, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return app.ErrInvalid
	case finfo.Size() < int64(req.FlowChunkSize):
		return nil
	case finfo.Size() == int64(req.FlowChunkSize):
		return app.ErrExists
	default:
		return app.ErrInvalid
	}
}

type mcdirRequestPath struct{}

func newMCDirRequestPath() *mcdirRequestPath {
	return &mcdirRequestPath{}
}

func (p *mcdirRequestPath) Path(req *flow.Request) string {
	return filepath.Join(p.Dir(req), fmt.Sprintf("%d", req.FlowChunkNumber))
}

func (p *mcdirRequestPath) Dir(req *flow.Request) string {
	mcdir := app.MCDir.Path()
	uploadPath := filepath.Join(mcdir, "upload", req.ProjectID, req.DirectoryID, req.FileID)
	return uploadPath
}

type nopRequestWriter struct {
	err error
}

func (r *nopRequestWriter) Write(req *flow.Request) error {
	return r.err
}
