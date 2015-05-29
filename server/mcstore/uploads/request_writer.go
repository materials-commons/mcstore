package uploads

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

// RequestWriter is the interface used to write a request.
type requestWriter interface {
	write(dir string, req *flow.Request) error
}

// A fileRequestWriter implements writing a request to a file.
type fileRequestWriter struct{}

// Write will write the blocks for a request to the path returned by
// the RequestPath Path call. Write will attempt to create the directory
// path to write to.
func (r *fileRequestWriter) write(dir string, req *flow.Request) error {
	path := filepath.Join(dir, fmt.Sprintf("%d", req.FlowChunkNumber))
	err := r.validateWrite(dir, path, req)
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
// If the size of the on disk chunk is smaller than the request
// chunk then that chunk is incomplete and we allow a write to it.
func (r *fileRequestWriter) validateWrite(dir, path string, req *flow.Request) error {
	// Create directory where chunk will be written
	if err := os.MkdirAll(dir, 0700); err != nil {
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
