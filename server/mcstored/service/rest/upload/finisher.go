package upload

import (
	"crypto/md5"
	"os"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// A FinisherFactory creates a new Finisher for a given uploadID and fileID.
type FinisherFactory interface {
	Finisher(req *flow.Request) Finisher
}

// A uploadFinisherFactory implements the actual finisher used by the upload service
// that performs cleanup and insertion of an uploaded file into the database.
type uploadFinisherFactory struct {
	tracker *uploadTracker
}

// NewUploadFinisherFactory creates a new uploadFinisherFactory.
func NewUploadFinisherFactory(tracker *uploadTracker) *uploadFinisherFactory {
	return &uploadFinisherFactory{
		tracker: tracker,
	}
}

// Finisher creates a new Finisher for the uploadFinisherFactory.
func (f *uploadFinisherFactory) Finisher(req *flow.Request) Finisher {
	return newUploadFinisher(req, f.tracker)
}

// A Finisher implements the method to call when assembly has finished successfully.
type Finisher interface {
	Finish() error
}

// uploadFinisher performs file cleanup, and database updates when a file has
// been successfully uploaded.
type uploadFinisher struct {
	tracker *uploadTracker
	req     *flow.Request
	files   dai.Files
}

// newUploadFinisher creates a new Finisher for the given uploadID and fileID. It uses the
// tracker to mark an upload as done by removing references to it.
func newUploadFinisher(req *flow.Request, tracker *uploadTracker) *uploadFinisher {
	return &uploadFinisher{
		tracker: tracker,
		req:     req,
	}
}

// Finish removes the temporary directory containing the chunks.
// It also clears the uploadID from the tracker since the upload
// has been completed and processed.
func (f *uploadFinisher) Finish() error {
	current, parent := f.isCurrent()
	fields := map[string]interface{}{
		schema.FileFields.Current():  current,
		schema.FileFields.Parent():   parent,
		schema.FileFields.Uploaded(): f.Size(),
	}
	checksum, err := file.HashStr(md5.New(), app.MCDir.FilePath(f.req.FileID))
	if err != nil {
		// log err and cleanup
		return err
	}
	fields[schema.FileFields.Checksum()] = checksum
	if matchingFile, found := f.findMatchingChecksum(checksum); found {
		// 1. Delete the uploaded file
		// 2. Insert an entry with usesid set to the matchingFile
		os.Remove(app.MCDir.FilePath(f.req.FileID))
		fields[schema.FileFields.UsesID()] = matchingFile.ID

	}
	f.files.UpdateFields(f.req.FileID, fields)
	f.tracker.clear(f.req.UploadID())
	os.RemoveAll(app.MCDir.UploadDir(f.req.UploadID()))
	return nil
}

func (f *uploadFinisher) findMatchingChecksum(checksum string) (matchingFile *schema.File, found bool) {
	matchingFile, err := f.files.ByChecksum(checksum)
	if err != nil {
		return nil, false
	}
	return matchingFile, true
}

func (f *uploadFinisher) isCurrent() (bool, string) {
	return false, ""
}

func (f *uploadFinisher) Size() int64 {
	finfo, err := os.Stat(app.MCDir.FilePath(f.req.FileID))
	if err != nil {
		return 0
	}
	return finfo.Size()
}
