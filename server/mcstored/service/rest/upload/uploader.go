package upload

import "github.com/materials-commons/mcstore/pkg/app/flow"

// TODO: Uploader will need to know the destination for assembly.
// TODO: Uploader will need a Finisher to be passed in to it.
// TODO: ***** To simplify the above, just pass in the assembler *****

type uploader struct {
	tracker       *uploadTracker
	requestWriter RequestWriter
	assembler     *Assembler
}

func NewUploader(requestWriter RequestWriter, tracker *uploadTracker) *uploader {
	return &uploader{
		tracker:       tracker,
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
