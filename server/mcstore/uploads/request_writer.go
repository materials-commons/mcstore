package uploads

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/materials-commons/gohandy/file"
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

// blockRequestWriter implements writing requests to a single file. It writes the
// requests in order by creating a sparse file and then seeking to the proper spot
// in the file to write the requests data.
type blockRequestWriter struct{}

// write will write the request to a file located in dir. The file will have
// the name of the flow UploadID(). This method creates a sparse file the
// size of the file to be written and then writes requests in order. Out of
// order chunks are handled by seeking to proper position in the file.
func (r *blockRequestWriter) write(dir string, req *flow.Request) error {
	path := filepath.Join(dir, req.UploadID())
	if err := r.validate(dir, path, req.FlowTotalSize); err != nil {
		return err
	}
	return r.writeRequest(path, req)
}

// writeRequest performs the actual write of the request. It opens the file
// sparse file, seeks to the proper position and then writes the data.
func (r *blockRequestWriter) writeRequest(path string, req *flow.Request) error {
	if f, err := os.OpenFile(path, os.O_WRONLY, 0660); err != nil {
		return err
	} else {
		defer f.Close()
		fromBeginning := 0
		seekTo := int64((req.FlowChunkNumber - 1) * int32(len(req.Chunk)))
		if _, err := f.Seek(seekTo, fromBeginning); err != nil {
			app.Log.Critf("Failed seeking to write chunk #%d for %s: %s", req.FlowChunkNumber, req.UploadID(), err)
			return err
		}

		if _, err := f.Write(req.Chunk); err != nil {
			app.Log.Critf("Failed writing chunk #%d for %s: %s", req.FlowChunkNumber, req.UploadID(), err)
			return err
		}
		return nil
	}
}

// validate ensures that the path exists. It will create the directory and
// the file. The file is created as a sparse file.
func (r *blockRequestWriter) validate(dir, path string, size int64) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if !file.Exists(path) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return f.Truncate(size)
	}
	return nil
}
