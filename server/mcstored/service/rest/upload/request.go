package upload

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

// RequestWriter is the interface used to write a request.
type RequestWriter interface {
	Write(req *flow.Request) error
}

// RequestPath is the interface for getting request locations.
type RequestPath interface {
	Path(req *flow.Request) string
	Dir(req *flow.Request) string
}

// A fileRequestWriter implements writing a request to a file.
type fileRequestWriter struct {
	RequestPath
}

// NewFileRequestWriter creates a new fileRequestWriter.
func NewFileRequestWriter(requestPath RequestPath) *fileRequestWriter {
	return &fileRequestWriter{
		RequestPath: requestPath,
	}
}

// Write will write the blocks for a request to the path returned by
// the RequestPath Path call. Write will attempt to create the directory
// path to write to.
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

// mcdirRequestPath implements RequestPath by returning paths starting
// from MCDIR. It gets the dir path from app.MCDir.Path().
type mcdirRequestPath struct{}

// NewMCDirRequestPath creates a new mcdirRequestPath.
func NewMCDirRequestPath() *mcdirRequestPath {
	return &mcdirRequestPath{}
}

// Path returns the full file path for a request. The path is constructed
// from app.MCDir.Path() and the request FlowChunkNumber. This allows the
// blocks for a file upload to be sorted so the file can be constructed.
func (p *mcdirRequestPath) Path(req *flow.Request) string {
	return filepath.Join(p.Dir(req), fmt.Sprintf("%d", req.FlowChunkNumber))
}

// Dir returns the path to put the request blocks to. The path is constructed
// from app.MCDir.Path and the request.UploadID().
func (p *mcdirRequestPath) Dir(req *flow.Request) string {
	mcdir := app.MCDir.Path()
	uploadPath := filepath.Join(mcdir, "upload", req.UploadID())
	return uploadPath
}

type nopRequestWriter struct {
	err error
}

func (r *nopRequestWriter) Write(req *flow.Request) error {
	return r.err
}
