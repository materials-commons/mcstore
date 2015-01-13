package app

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/config"
)

type mcdir struct{}

// MCDir gives access to methods on the MCDIR directory. It will
// panic if MCDIR is not set.
var MCDir mcdir

func (d mcdir) Path() string {
	dir := config.GetString("MCDIR")
	if dir == "" {
		panic("MCDIR not set")
	}
	return dir
}

// FileDir returns the directory path for a given fileID. It returns the
// empty string if a bad fileID is given.
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

// FileConversionDir returns the conversion directory path for a fileID. The
// conversion directory is the directory where converted image files are kept.
// It returns the empty string if a bad fileID is given.
func (d mcdir) FileConversionDir(fileID string) string {
	dir := d.FileDir(fileID)
	return makePath(dir, ".conversion")
}

// FilePath returns the full path including the file for a fileID.
// It returns the empty string if a bad fileID is given.
func (d mcdir) FilePath(fileID string) string {
	dir := d.FileDir(fileID)
	return makePath(dir, fileID)
}

// FilePathImageConversion returns the full path, include the file,
// to the converted image file. It returns the empty string if a
// bad fileID is given.
func (d mcdir) FilePathImageConversion(fileID string) string {
	dir := d.FileConversionDir(fileID)
	return makePath(dir, fileID+".jpg")
}

// makePath constructs the file path. It handles checking
// for bad segments and empty directory paths. It returns
// the empty string if the first segment is empty, or
// if the list of segments is empty.
func makePath(pathSegments ...string) string {
	switch {
	case len(pathSegments) == 0:
		return ""
	case pathSegments[0] == "":
		return ""
	default:
		return filepath.Join(pathSegments...)
	}
}
