package upload

import (
	"os"

	"github.com/materials-commons/mcstore/pkg/app"
)

// A FinisherFactory creates a new Finisher for a given uploadID and fileID.
type FinisherFactory interface {
	Finisher(uploadID, fileID string) Finisher
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
func (f *uploadFinisherFactory) Finisher(uploadID, fileID string) Finisher {
	return newUploadFinisher(uploadID, f.tracker, fileID)
}

// A Finisher implements the method to call when assembly has finished successfully.
type Finisher interface {
	Finish() error
}

// uploadFinisher performs file cleanup, and database updates when a file has
// been successfully uploaded.
type uploadFinisher struct {
	uploadID string
	tracker  *uploadTracker
	fileID   string
}

// newUploadFinisher creates a new Finisher for the given uploadID and fileID. It uses the
// tracker to mark an upload as done by removing references to it.
func newUploadFinisher(uploadID string, tracker *uploadTracker, fileID string) *uploadFinisher {
	return &uploadFinisher{
		uploadID: uploadID,
		tracker:  tracker,
		fileID:   fileID,
	}
}

func (f *uploadFinisher) Finish() error {
	f.tracker.clear(f.uploadID)
	os.RemoveAll(app.MCDir.UploadDir(f.uploadID))
	return nil
}
