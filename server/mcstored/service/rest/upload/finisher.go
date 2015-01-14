package upload

import (
	"os"

	"github.com/materials-commons/mcstore/pkg/app"
)

type FinisherFactory interface {
	Finisher(uploadID, fileID string) Finisher
}

type uploadFinisherFactory struct {
	tracker *uploadTracker
}

func NewUploadFinisherFactory(tracker *uploadTracker) *uploadFinisherFactory {
	return &uploadFinisherFactory{
		tracker: tracker,
	}
}

func (f *uploadFinisherFactory) Finisher(uploadID, fileID string) Finisher {
	return newUploadFinisher(uploadID, f.tracker, fileID)
}

// A Finisher implements the method to call when assembly has finished successfully.
type Finisher interface {
	Finish() error
}

type uploadFinisher struct {
	uploadID string
	tracker  *uploadTracker
	fileID   string
}

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
