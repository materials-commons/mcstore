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
	return config.GetString("MCDIR")
}

func (d mcdir) FileDir(fileID string) string {
	idSegments := strings.Split(fileID, "-")
	return filepath.Join(d.Path(), idSegments[1][0:2], idSegments[1][2:4])
}

func (d mcdir) FilePath(fileID string) string {
	return filepath.Join(d.FileDir(fileID), fileID)
}
