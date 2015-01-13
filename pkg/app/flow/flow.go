package flow

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
)

const ChunkPerms = 0700

// A FlowRequest encapsulates the flowjs protocol for uploading a file. The
// protocol supports extensions to the protocol. We extend the protocol to
// include Materials Commons specific information. It is also expected that
// the data sent by flow or another client will be placed in chunkData.
type Request struct {
	FlowChunkNumber  int32  `json:"flowChunkNumber"`  // The chunk being sent.
	FlowTotalChunks  int32  `json:"flowTotalChunks"`  // The total number of chunks to send.
	FlowChunkSize    int32  `json:"flowChunkSize"`    // The size of the chunk.
	FlowTotalSize    int64  `json:"flowTotalSize"`    // The size of the file being uploaded.
	FlowIdentifier   string `json:"flowIdentifier"`   // A unique identifier used by Flow. Not guaranteed to be a GUID.
	FlowFileName     string `json:"flowFilename"`     // The file name being uploaded.
	FlowRelativePath string `json:"flowRelativePath"` // When available the relative file path.
	ProjectID        string `json:"projectID"`        // Materials Commons Project ID.
	DirectoryID      string `json:"directoryID"`      // Materials Commons Directory ID.
	FileID           string `json:"fileID"`           // Materials Commons File ID.
	Chunk            []byte `json:"-"`                // The file data.
	ChunkHash        string `json:"chunkHash"`        // The computed MD5 hash for the chunk (optional).
	FileHash         string `json:"fileHash"`         // The computed MD5 hash for the file (optional)
}

func (r *Request) UploadID() string {
	return fmt.Sprintf("%s-%s-%s", r.ProjectID, r.DirectoryID, r.FileID)
}

func (r *Request) Write() error {
	path := r.Path()
	err := r.validateWrite(path)
	switch {
	case err == nil:
		return ioutil.WriteFile(path, r.Chunk, ChunkPerms)
	case err == app.ErrExists:
		return nil
	default:
		return err
	}
}

// validateWrite determines if a particular chunk can be written.
// If the size of the ondisk chunk is smaller than the request
// chunk then that chunk is incomplete and we allow a write to it.
func (r *Request) validateWrite(path string) error {
	if err := os.MkdirAll(path, ChunkPerms); err != nil {
		return err
	}
	finfo, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return app.ErrInvalid
	case finfo.Size() < int64(r.FlowChunkSize):
		return nil
	case finfo.Size() == int64(r.FlowChunkSize):
		return app.ErrExists
	default:
		return app.ErrInvalid
	}
}

func (r *Request) Path() string {
	return filepath.Join(r.Dir(), fmt.Sprintf("%d", r.FlowChunkNumber))
}

func (r *Request) Dir() string {
	mcdir := app.MCDir.Path()
	uploadPath := filepath.Join(mcdir, "upload", r.ProjectID, r.DirectoryID, r.FileID)
	return uploadPath
}
