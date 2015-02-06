package file

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// darwinExFileInfo stores windows specific file information.
// At the moment all the information we need is available
// through the Sys() interface.
type darwinExFileInfo struct {
	os.FileInfo
	fid  FID
	path string
}

// timespecToTime converts a unix timespec into a time.Time. This was
// copied from os/stat_linux.go in the Go source.
func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

// CTime returns the creation time (ctime) from stat_t.
func (fi *darwinExFileInfo) CTime() time.Time {
	return timespecToTime(fi.Sys().(*syscall.Stat_t).Ctimespec)
}

// ATime returns the access time (atime) from stat_t
func (fi *darwinExFileInfo) ATime() time.Time {
	return timespecToTime(fi.Sys().(*syscall.Stat_t).Atimespec)
}

// FID returns the file id based on the inode.
func (fi *darwinExFileInfo) FID() FID {
	return fi.fid
}

// Path returns the full path for the file
func (fi *darwinExFileInfo) Path() string {
	return fi.path
}

// newExFileInfo creates a new winExFileInfo from a os.FileInfo.
func newExFileInfo(fi os.FileInfo, path string) *darwinExFileInfo {
	fid := FID{
		IDLow: fi.Sys().(*syscall.Stat_t).Ino,
	}
	absolute, _ := filepath.Abs(path)
	return &darwinExFileInfo{
		FileInfo: fi,
		fid:      fid,
		path:     filepath.Clean(absolute),
	}
}
