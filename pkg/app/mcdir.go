package app

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/file"
)

type mcdir struct{}

// MCDir gives access to methods on the MCDIR directory. It will
// panic if MCDIR is not set.
var MCDir mcdir

func (d mcdir) Path() string {
	dirs := config.GetString("MCDIR")
	if dirs == "" {
		panic("MCDIR not set")
	}
	return strings.Split(dirs, ":")[0]
}

func (d mcdir) Paths() []string {
	dirs := config.GetString("MCDIR")
	if dirs == "" {
		panic("MCDIR not set")
	}

	return strings.Split(dirs, ":")
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

func (d mcdir) FileDirFromPath(path, fileID string) string {
	idSegments := strings.Split(fileID, "-")
	switch {
	case len(idSegments) < 2:
		Log.Debug(Logf("FileDir: Got bad fileID: %s", fileID))
		return ""
	case len(idSegments[1]) < 4:
		Log.Debug(Logf("FileDir: Got bad fileID: %s", fileID))
		return ""
	default:
		return filepath.Join(path, idSegments[1][0:2], idSegments[1][2:4])
	}
}

// FileConversionDir returns the conversion directory path for a fileID. The
// conversion directory is the directory where converted image files are kept.
// It returns the empty string if a bad fileID is given.
func (d mcdir) FileConversionDir(fileID string) string {
	dir := d.FileDir(fileID)
	return makePath(dir, ".conversion")
}

func (d mcdir) FileConversionDirFromPath(dirPath, fileID string) string {
	dir := d.FileDirFromPath(dirPath, fileID)
	return makePath(dir, ".conversion")
}

// FilePath returns the full path including the file for a fileID.
// It returns the empty string if a bad fileID is given.
func (d mcdir) FilePath(fileID string) string {
	for _, dirPath := range d.Paths() {
		fileDirPath := d.FileDirFromPath(dirPath, fileID)
		filePath := makePath(fileDirPath, fileID)
		if file.Exists(filePath) {
			return filePath
		}
	}

	// By default return path from first entry if loop fails
	dir := d.FileDir(fileID)
	return makePath(dir, fileID)
}

// FilePathImageConversion returns the full path, include the file,
// to the converted image file. It returns the empty string if a
// bad fileID is given.
func (d mcdir) FilePathImageConversion(fileID string) string {
	for _, dirPath := range d.Paths() {
		fileDirPath := d.FileConversionDirFromPath(dirPath, fileID)
		filePath := makePath(fileDirPath, fileID+".jpg")
		if file.Exists(filePath) {
			return filePath
		}
	}

	// By default return path from first entry if loop fails
	dir := d.FileConversionDir(fileID)
	return makePath(dir, fileID+".jpg")
}

// FilePathFromConversionToPDF returns the full path, include the file,
// to the converted image file. It returns the empty string if a
// bad fileID is given.
func (d mcdir) FilePathFromConversionToPDF(fileID string) string {
	for _, dirPath := range d.Paths() {
		fileDirPath := d.FileConversionDirFromPath(dirPath, fileID)
		filePath := makePath(fileDirPath, fileID+".pdf")
		if file.Exists(filePath) {
			return filePath
		}
	}

	// By default return path from first entry if loop fails
	dir := d.FileConversionDir(fileID)
	return makePath(dir, fileID+".pdf")
}

// UploadDir returns the path to the upload directory for a
// given uploadID.
func (d mcdir) UploadDir(uploadID string) string {
	return filepath.Join(d.Path(), "upload", uploadID)
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
