package uploads

import (
	"crypto/md5"
	"os"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type finisher struct {
	files dai.Files
}

func newFinisher(files dai.Files) *finisher {
	return &finisher{
		files: files,
	}
}

// TODO: refactor this method into a few separate methods that contain the
// logical blocks.
func (f *finisher) finish(req *UploadRequest, fileID, dirID string) error {
	checksum, err := file.HashStr(md5.New(), app.MCDir.FilePath(fileID))
	if err != nil {
		// log
		return err
	}

	parentID, err := f.parentID(req.FlowFileName, dirID)
	if err != nil {
		app.Log.Errorf("Looking up parentID for %s/%s returned unexpected error: %s.", req.FlowFileName, dirID, err)
		return err
	}

	size := f.Size(fileID)

	if size != req.FlowTotalSize {
		app.Log.Errorf("Uploaded file (%s/%s) doesn't match the expected size. Expected:%d, Got: %d", req.FlowFileName, req.FlowIdentifier, req.FlowTotalSize, size)
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
		// Found a matching checksum so set file entry
		// to point at this file and remove the file that
		// was just constructed.
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
