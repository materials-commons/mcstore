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

func (d mcdir) FileIDDir(fileID string) string {
	pieces := strings.Split(fileID, "-")
	return filepath.Join(d.Path(), pieces[1][0:2])
}

func (d mcdir) FileIDPath(fileID string) string {
	return filepath.Join(d.FileIDDir(fileID), fileID)
}
