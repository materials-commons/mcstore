package uploads

import (
	"fmt"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

// RequestPath is the interface for getting request locations.
type RequestPath interface {
	Path(req *flow.Request) string
	Dir(req *flow.Request) string
}

// mcdirRequestPath implements RequestPath by returning paths starting
// from MCDIR. It gets the dir path from app.MCDir.Path().
type mcdirRequestPath struct{}

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
