package uploads

import (
	"crypto/md5"
	"fmt"
	"os"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type finisher struct {
	files   dai.Files
	uploads dai.Uploads
}

func newFinisher(files dai.Files, uploads dai.Uploads) *finisher {
	return &finisher{
		files:   files,
		uploads: uploads,
	}
}

// TODO: refactor this method into a few separate methods that contain the
// logical blocks.
func (f *finisher) finish(req *UploadRequest, fileID string, upload *schema.Upload) error {
	checksum, err := file.HashStr(md5.New(), app.MCDir.FilePath(fileID))
	if err != nil {
		// log
		return err
	}

	parentID, err := f.parentID(upload.File.Name, upload.DirectoryID)
	if err != nil {
		app.Log.Errorf("Looking up parentID for %s/%s returned unexpected error: %s.", upload.File.Name, upload.DirectoryID, err)
		return err
	}

	size := f.Size(fileID)

	if size != req.FlowTotalSize {
		app.Log.Errorf("Uploaded file (%s/%s) doesn't match the expected size. Expected:%d, Got: %d", upload.File.Name, req.FlowIdentifier, req.FlowTotalSize, size)
		return app.ErrInvalid
	}

	fields := map[string]interface{}{
		schema.FileFields.Current():  true,
		schema.FileFields.Parent():   parentID,
		schema.FileFields.Uploaded(): req.FlowTotalSize,
		schema.FileFields.Size():     req.FlowTotalSize,
		schema.FileFields.Checksum(): checksum,
	}

	matchingFile, err := f.files.ByChecksum(checksum)
	switch {
	case err != nil && err == app.ErrNotFound:
		// Nothing to do, we already set the checksum and
		// usesid is clear.
	case err != nil:
		// Some type of error accessing the database
		app.Log.Errorf("Looking up file by checksum for %s/%s returned unexpected error: %s.", checksum, req.FlowFileName, err)
		return err
	default:
		// Found a matching checksum. There are two cases
		// 1. The existing file is not the file we uploaded
		// 2. The existing file is the file we uploaded.

		// Is matching file, delete it completely.
		if f.isSameFile(matchingFile, upload.File.Name, upload.DirectoryID) {
			app.Log.Infof("Found exact matching file (%s/%s) in same directory (%s), deleting", upload.File.Name, fileID, upload.DirectoryID)
			f.deleteUploadedFile(fileID, upload)
			return nil
		}

		// Not matching file so set file entry to point
		// at this file and remove the file that was
		// just constructed.

		fields[schema.FileFields.UsesID()] = matchingFile.ID
		os.Remove(app.MCDir.FilePath(fileID))
	}

	return f.files.UpdateFields(fileID, fields)
}

func (f *finisher) Size(fileID string) int64 {
	finfo, err := os.Stat(app.MCDir.FilePath(fileID))
	if err != nil {
		return 0
	}
	return finfo.Size()
}

func (f *finisher) parentID(fileName, dirID string) (parentID string, err error) {
	parent, err := f.files.ByPath(fileName, dirID)

	if parent != nil {
		parentID = parent.ID
	}

	if err != nil && err == app.ErrNotFound {
		// log error
		return parentID, nil
	}

	return parentID, err
}

func (f *finisher) isSameFile(matchingFile *schema.File, fileName, dirID string) bool {
	fmt.Printf("isSameFile %#v, %s, %s\n", matchingFile, fileName, dirID)
	if matchingFile.Name != fileName {
		fmt.Println("isSameFile names don't match")
		return false
	}

	dirs, err := f.files.Directories(matchingFile.ID)
	fmt.Printf("isSameFile dirs %#v\n", dirs)
	if err != nil {
		return false
	}
	for _, entryID := range dirs {
		if entryID == dirID {
			return true
		}
	}
	return false
}

func (f *finisher) deleteUploadedFile(fileID string, upload *schema.Upload) {
	f.files.Delete(fileID, upload.DirectoryID, upload.ProjectID)
	os.Remove(app.MCDir.FilePath(fileID))
}
