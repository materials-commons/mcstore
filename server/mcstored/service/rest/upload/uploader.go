package upload

import (
	"os"

	"github.com/materials-commons/mcstore/pkg/app/flow"
)

// TODO: Uploader will need to know the destination for assembly.
// TODO: Uploader will need a Finisher to be passed in to it.
// TODO: ***** To simplify the above, just pass in the assembler *****

type uploader struct {
	tracker       *uploadTracker
	requestWriter RequestWriter
	assembler     *Assembler
}

func newUploader(requestWriter RequestWriter) *uploader {
	return &uploader{
		tracker:       newUploadTracker(),
		requestWriter: requestWriter,
	}
}

func (u *uploader) processRequest(request *flow.Request) error {
	if err := u.requestWriter.Write(request); err != nil {
		// write failed for some reason
		return err
	}

	// Increment block count
	id := request.UploadID()
	u.tracker.increment(id)
	return nil
}

func (u *uploader) allBlocksUploaded(request *flow.Request) bool {
	id := request.UploadID()
	count := u.tracker.count(id)
	return count == request.FlowTotalChunks
}

type uploadFinisher struct {
	uploadID  string
	tracker   *uploadTracker
	uploadDir string
}

func newUploadFinisher(uploadID string, tracker *uploadTracker, uploadDir string) *uploadFinisher {
	return &uploadFinisher{
		uploadID:  uploadID,
		tracker:   tracker,
		uploadDir: uploadDir,
	}
}

func (f *uploadFinisher) Finish() error {
	f.tracker.clear(f.uploadID)
	os.RemoveAll(f.uploadDir)
	return nil
}
