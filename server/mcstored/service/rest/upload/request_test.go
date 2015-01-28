package upload

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/stretchr/testify/require"
)

func TestMCDirRequestPathPath(t *testing.T) {
	r := NewMCDirRequestPath()
	// For this test we only care about the chunk number
	// and the FlowIdentifier in the flow request
	req := &flow.Request{
		FlowChunkNumber: 1,
		FlowIdentifier:  "abc",
	}
	mcdir := app.MCDir.Path()
	expectedPath := filepath.Join(mcdir, "upload", req.FlowIdentifier, fmt.Sprintf("%d", req.FlowChunkNumber))
	require.Equal(t, expectedPath, r.Path(req))
}

func TestMCDirRequestPathDir(t *testing.T) {
	r := NewMCDirRequestPath()
	// For this test we only care about the chunk number
	// and the FlowIdentifier in the flow request
	req := &flow.Request{
		FlowChunkNumber: 1,
		FlowIdentifier:  "abc",
	}
	mcdir := app.MCDir.Path()
	expectedDir := filepath.Join(mcdir, "upload", req.FlowIdentifier)
	require.Equal(t, expectedDir, r.Dir(req))
}
