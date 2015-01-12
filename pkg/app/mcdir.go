package app

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/config"
)

type mcdir struct{}

// MCDir gives access to methods on the MCDIR directory.
var MCDir mcdir

func (d mcdir) Path() string {
	dir := config.GetString("MCDIR")
	if dir == "" {
		panic("MCDIR not set")
	}
	return dir
}

func (d mcdir) FileDir(fileID string) string {
	idSegments := strings.Split(fileID, "-")
	switch {
	case len(idSegments) < 2:
		Log.Debug(Logf("FileDir: Got bad fileID: %s", fileID))
		return ""
	case len(idSegments[1]) < 4:
		Log.Debug(Logf("FileDir: Got bad fileID: %s", fileID))
		return ""
	default:
		return filepath.Join(d.Path(), idSegments[1][0:2], idSegments[1][2:4])
	}
}

func (d mcdir) FilePath(fileID string) string {
	dir := d.FileDir(fileID)
	if dir == "" {
		return ""
	}
	return filepath.Join(d.FileDir(fileID), fileID)
}
