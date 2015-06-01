package uploads

import (
	"fmt"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"os"
)

// RequestPath is the interface for getting request locations.
type requestPath interface {
	path(req *flow.Request) string
	dir(req *flow.Request) string
	dirFromID(id string) string
	mkdir(req *flow.Request) error
	mkdirFromID(id string) error
}

// mcdirRequestPath implements RequestPath by returning paths starting
// from MCDIR. It gets the dir path from app.MCDir.Path().
type mcdirRequestPath struct{}

// path returns the full file path for a request. The path is constructed
// from app.MCDir.Path() and the request FlowChunkNumber. This allows the
// blocks for a file upload to be sorted so the file can be constructed.
func (p *mcdirRequestPath) path(req *flow.Request) string {
	return filepath.Join(p.dir(req), fmt.Sprintf("%d", req.FlowChunkNumber))
}

// dir returns the path to put the request blocks to. The path is constructed
// from app.MCDir.Path and the request.UploadID().
func (p *mcdirRequestPath) dir(req *flow.Request) string {
	return p.dirFromID(req.UploadID())
}

func (p *mcdirRequestPath) dirFromID(id string) string {
	mcdir := app.MCDir.Path()
	uploadPath := filepath.Join(mcdir, "upload", id)
	return uploadPath
}

// mkdir creates the directory for the request.
func (p *mcdirRequestPath) mkdir(req *flow.Request) error {
	return p.mkdirFromID(req.UploadID())
}

// mkdirFromID creates the directory for the request ID. The path to the directory
// is constructed from the dirFromID call.
func (p *mcdirRequestPath) mkdirFromID(id string) error {
	mcdir := app.MCDir.Path()
	path := filepath.Join(mcdir, "upload", id)
	return os.MkdirAll(path, 0777)
}
