package uploads

import (
	"crypto/md5"
	"os"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstore/uploads/processor"
)

// A finisher finishes the rest of the book keeping
// when a file has been successfully uploaded and
// reconstructed.
type finisher struct {
	files dai.Files
	dirs  dai.Dirs
}

// newFinisher creates a new finisher.
func newFinisher(files dai.Files, dirs dai.Dirs) *finisher {
	return &finisher{
		files: files,
		dirs:  dirs,
	}
}

// finish takes care of updating the file and directory pointers, determining
// if a matching file (by checksum) has already been uploaded, and making the
// file ready for the user to access.
// TODO: refactor this method into a few separate methods that contain the
// logical blocks.
func (f *finisher) finish(req *UploadRequest, fileID string, upload *schema.Upload) error {
	filePath := app.MCDir.FilePath(fileID)
	checksum, err := file.HashStr(md5.New(), filePath)
	if err != nil {
		app.Log.Errorf("Failed creating checksum for '%s': %s", filePath, err)
		return err
	}

	parentID, err := f.parentID(upload.File.Name, upload.DirectoryID)
	if err != nil {
		app.Log.Errorf("Looking up parentID for %s/%s returned unexpected error: %s.", upload.File.Name, upload.DirectoryID, err)
		return err
	}

	size := f.size(fileID)

	if size != req.FlowTotalSize {
		app.Log.Errorf("Uploaded file (%s/%s) doesn't match the expected size. Expected:%d, Got: %d", upload.File.Name, req.FlowIdentifier, req.FlowTotalSize, size)
		return app.ErrInvalid
	}

	mediatype := files.MediaType(upload.File.Name, filePath)
	fields := map[string]interface{}{
		schema.FileFields.Current():   true,
		schema.FileFields.Parent():    parentID,
		schema.FileFields.Uploaded():  req.FlowTotalSize,
		schema.FileFields.Size():      req.FlowTotalSize,
		schema.FileFields.Checksum():  checksum,
		schema.FileFields.MediaType(): mediatype,
	}

	matchingFile, err := f.files.ByChecksum(checksum)
	switch {
	case err != nil && err == app.ErrNotFound:
		// This is a brand new upload for a file we haven't seen before. There are processing
		// steps that may need to be done on the file. For example we convert tif and bmp
		// image files so they can be displayed in the browser.
		f.processFile(fileID, mediatype)
	case err != nil:
		// Some type of error accessing the database
		app.Log.Errorf("Looking up file by checksum for %s/%s returned unexpected error: %s.", checksum, req.FlowFileName, err)
		return err
	default:
		// Found a matching checksum. There are two cases
		// 1. The existing file is the file we uploaded.
		// 2. The existing file is not the file we uploaded

		// Case 1: Is matching file. Delete it completely.
		if f.fileInDir(checksum, upload.File.Name, upload.DirectoryID) {
			app.Log.Infof("Found exact matching file (%s/%s) in same directory (%s), deleting", upload.File.Name, fileID, upload.DirectoryID)
			f.deleteUploadedFile(fileID, upload)
			return nil
		}

		// Case 2: Existing file is not the file we uploaded:
		// File found is not the file we uploaded, so we need to do a couple of things:
		// First set file entry to point this file (set UsesID) and secondly
		// remove the file (on the file system) that was just uploaded since we don't
		// keep duplicates.

		fields[schema.FileFields.UsesID()] = matchingFile.ID
		os.Remove(filePath)
	}

	return f.files.UpdateFields(fileID, fields)
}

// Size gets the size of the reconstructed file.
func (f *finisher) size(fileID string) int64 {
	finfo, err := os.Stat(app.MCDir.FilePath(fileID))
	if err != nil {
		return 0
	}
	return finfo.Size()
}

// parentID returns the parent for this file (if any).
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

// processFile will process the file on disk.
func (f *finisher) processFile(fileID string, mediatype schema.MediaType) {
	fp := processor.New(fileID, mediatype)
	fp.Process()
}

// fileInDir determines if this exact file has already been uploaded
// to this directory.
func (f *finisher) fileInDir(checksum, fileName, dirID string) bool {
	files, err := f.dirs.Files(dirID)
	if err != nil {
		return false
	}

	for _, fileEntry := range files {
		if fileEntry.Name == fileName && fileEntry.Checksum == checksum {
			return true
		}
	}
	return false
}

// deleteUploadedFile will completely delete the file from the system. Completely
// means it also deletes the database entries for the file.
func (f *finisher) deleteUploadedFile(fileID string, upload *schema.Upload) {
	f.files.Delete(fileID, upload.DirectoryID, upload.ProjectID)
	os.Remove(app.MCDir.FilePath(fileID))
}
