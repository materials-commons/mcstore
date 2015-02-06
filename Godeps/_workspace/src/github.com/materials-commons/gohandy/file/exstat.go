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

	exfi := newExFileInfo(fi, path)
	return exfi, nil
}
