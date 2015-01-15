package upload

import "github.com/materials-commons/mcstore/pkg/app/flow"

// uploader handles uploading the blocks for a request. It tracks
// the blocks that have been uploaded so that its users can query
// whether an upload has finished.
type uploader struct {
	tracker       *uploadTracker // tracker used to track block count uploaded.
	requestWriter RequestWriter  // where to write a request
}

// NewUploader creates a new uploader instance.
func NewUploader(requestWriter RequestWriter, tracker *uploadTracker) *uploader {
	return &uploader{
		tracker:       tracker,
		requestWriter: requestWriter,
	}
}

// processRequest will attempt to write a request to the RequestWriter. If
// successful it will increment the number of blocks uploaded. If all blocks
// have been uploaded, then processRequest will return nil and not attempt
// to write the block.
func (u *uploader) processRequest(request *flow.Request) error {
	if u.allBlocksUploaded(request) {
		return nil
	}

	if err := u.requestWriter.Write(request); err != nil {
		// write failed for some reason
		return err
	}

	// Increment block count
	id := request.UploadID()
	u.tracker.increment(id)
	return nil
}

// allBlocksUploaded will compare the number of blocks expected
// with the number of blocks uploaded. If they match it will
// return true (all blocks have been uploaded).
func (u *uploader) allBlocksUploaded(request *flow.Request) bool {
	id := request.UploadID()
	count := u.tracker.count(id)
	return count == request.FlowTotalChunks
}
