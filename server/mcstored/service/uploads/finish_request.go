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
	files       dai.Files
	requestPath RequestPath
}

func newFinisher(files dai.Files) *finisher {
	return &finisher{
		files: files,
	}
}

func (f *finisher) finish(req *UploadRequest, fileID, dirID string) error {
	checksum, err := file.HashStr(md5.New(), app.MCDir.FilePath(fileID))
	if err != nil {
		// log
		return err
	}

	parent, err := f.files.ByPath(req.FlowFileName, dirID)

	if err != nil && err != app.ErrNotFound {
		// log error
		return err
	}

	parentID := ""
	if parent != nil {
		parentID = parent.ID
	}

	// Should stat the file and make sure FlowTotalSize == fi.Size()

	size := f.Size(fileID)

	if size != req.FlowTotalSize {
		// log
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
		// Only set the checksum and not uses id
	case err != nil:
		// Some type of error accessing the database
		// log error
		return err
	default:
		// Found a matching checksum
		fields[schema.FileFields.UsesID()] = matchingFile.ID
		os.Remove(app.MCDir.FilePath(fileID)) // Remove uploaded file
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
