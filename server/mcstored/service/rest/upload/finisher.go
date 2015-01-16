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

/*
Finish code:
    Get the checksum
    if another file has that checksum then
       delete file just uploaded
       set usesid to that files checksum
    end
    Check if this file will be current file
    Get this files size
    Get this files parent
    Create file and insert into project and directory
    if there wasn't a matching checksum then
        move assembled file from upload area to app.MCDir.FilePath(fileID)
    end

    Other implied changes:
       * The assembler will assemble the file in the directory
         with the parts
       * There is no FileID in flow.Request
       * An upload request first needs to get an UploadID
       * We should track UploadIDs for a directory/machine/lastModifiedDate/Size combination
            - This will allow us to track incomplete uploads
            - It will also allow us check if the file has changed
            - It isn't perfect, but it is better than nothing
            - Better method is to compute a checksum for each block
            - we upload and compare that to the checksum for blocks
              already loaded

*/

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
