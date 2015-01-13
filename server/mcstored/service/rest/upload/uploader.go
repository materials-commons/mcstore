package upload

import "github.com/materials-commons/mcstore/pkg/app/flow"

type uploader struct {
	tracker *uploadTracker
	w       RequestWriter
}

func newUploader(w RequestWriter) *uploader {
	return &uploader{
		tracker: newUploadTracker(),
		w:       w,
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
	assembler := newAssembler(request, u.tracker)
	assembler.launch()
}

type uploadFinisher struct {
	uploadID string
	tracker  *uploadTracker
}

func newUploadFinisher(uploadID string, tracker *uploadTracker) *uploadFinisher {
	return &uploadFinisher{
		uploadID: uploadID,
		tracker:  tracker,
	}
}

func (f *uploadFinisher) Finish() error {
	f.tracker.clear(f.uploadID)
	return nil
}
