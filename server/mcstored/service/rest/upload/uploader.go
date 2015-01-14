package upload

import (
	"io"
	"os"

	"github.com/materials-commons/mcstore/pkg/app/flow"
)

// TODO: Uploader will need to know the destination for assembly.
// TODO: Uploader will need a Finisher to be passed in to it.
// TODO: ***** To simplify the above, just pass in the assembler *****

type uploader struct {
	tracker     *uploadTracker
	w           RequestWriter
	finisher    Finisher
	destination io.Writer
}

func newUploader(w RequestWriter, destination io.Writer, finisher Finisher) *uploader {
	return &uploader{
		tracker:     newUploadTracker(),
		w:           w,
		finisher:    finisher,
		destination: destination,
	}
}

func (u *uploader) processRequest(request *flow.Request) error {
	if err := u.w.Write(request); err != nil {
		// write failed for some reason
		return err
	}

	if u.uploadDone(request) {
		u.assembleUpload(request)
	}
	return nil
}

func (u *uploader) uploadDone(request *flow.Request) bool {
	id := request.UploadID()
	count := u.tracker.increment(id)
	return count == request.FlowTotalChunks
}

func (u *uploader) assembleUpload(request *flow.Request) {
	assembler := NewAssembler(nil, u.finisher) // Need list of Items
	go func() {
		assembler.To(u.destination) // fix this with real destination
	}()
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
