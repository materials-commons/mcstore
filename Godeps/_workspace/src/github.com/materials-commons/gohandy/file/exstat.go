package file

import (
	"os"
	"time"
)

// ExFileInfo is an extended version of the os.FileInfo interface that
// includes additional information.
type ExFileInfo interface {
	os.FileInfo       // Support the os.FileInfo interface
	Path() string     // Full path of file
	CTime() time.Time // Creation time
	ATime() time.Time // Last access time
	FID() FID         // System independent File ID
}

// ExStat is an extended version of the os.Stat() method.
func ExStat(path string) (fileInfo ExFileInfo, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	exfi := systemExFileInfo(fi, path)
	return exfi, nil
}

// ExStatFromFileInfo takes a os.FileInfo and a path and returns an ExFileInfo.
func ExStatFromFileInfo(fi os.FileInfo, path string) (ExFileInfo, error) {
	return systemExFileInfo(fi, path), nil
}

// stdExFileInfo is a version of the ExFileInfo where the extended
// attributes are given by the user.
type stdExFileInfo struct {
	os.FileInfo
	path  string
	mtime time.Time
	ctime time.Time
	atime time.Time
	fid   FID
	size  int64
}

func ExInfoFrom(size int64, ctime, atime, mtime time.Time, fid FID) *stdExFileInfo {
	finfo := &stdExFileInfo{
		path:  "",
		ctime: ctime,
		atime: atime,
		mtime: mtime,
		fid:   fid,
		size:  size,
	}
	return finfo
}

func (fi *stdExFileInfo) FID() FID {
	return fi.fid
}

func (fi *stdExFileInfo) CTime() time.Time {
	return fi.ctime
}

func (fi *stdExFileInfo) ModTime() time.Time {
	return fi.mtime
}

func (fi *stdExFileInfo) ATime() time.Time {
	return fi.atime
}

func (fi *stdExFileInfo) Path() string {
	return fi.path
}

func (fi *stdExFileInfo) Size() int64 {
	return fi.size
}
