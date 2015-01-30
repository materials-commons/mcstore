package uploads

import (
	"io"
	"os"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// A UploadRequest contains the block to upload and the
// information required to write that block.
type UploadRequest struct {
	*flow.Request
}

// UploadService takes care of uploading blocks and constructing the
// file when all blocks have been uploaded.
type UploadService interface {
	Upload(req *UploadRequest) error
}

// uploadService is an implementation of UploadService.
type uploadService struct {
	tracker *uploadTracker
	files   dai.Files
	uploads dai.Uploads
	dirs    dai.Dirs
}

// NewUploadService creates a new uploadService connecting
// to the default database. It will panic if it cannot
// establish a connection to the database.
func NewUploadService() *uploadService {
	session := db.RSessionMust()
	return &uploadService{
		tracker: newUploadTracker(),
		files:   dai.NewRFiles(session),
		uploads: dai.NewRUploads(session),
		dirs:    dai.NewRDirs(session),
	}
}

// Upload takes care of uploading a block and constructing the file
// after all blocks have been uploaded. It takes care of all the
// details such as files that have already been uploaded.
func (s *uploadService) Upload(req *UploadRequest) error {
	dir := s.requestDir(req.Request)

	if err := s.Write(dir, req.Request); err != nil {
		return err
	}

	id := req.UploadID()
	s.tracker.increment(id)

	if s.allBlocksUploaded(id, req.FlowTotalChunks) {
		if file, err := s.assemble(req, dir); err != nil {
			// assembly failed. If file isn't nil then
			// there is some cleanup to do in the database.
			if file != nil {
				err2 := s.cleanup(req, file.ID)
				app.Log.Errorf("Assembly failed for uploaded file: attempted cleanup of database entry returned: %s", err2)
			}
			return err
		}
	}

	return nil
}

// allBlocksUploaded checks if we have received all the blocks for a file.
func (s *uploadService) allBlocksUploaded(id string, totalChunks int32) bool {
	count := s.tracker.count(id)
	return count == totalChunks
}

// assemble put the chunks for the file back together, create a database entry
// and take care of all book keeping tasks to make the file accessible.
func (s *uploadService) assemble(req *UploadRequest, dir string) (*schema.File, error) {
	// Look up the upload
	upload, err := s.uploads.ByID(req.FlowIdentifier)
	if err != nil {
		return nil, err
	}

	// Create file entry in database
	file, err := s.createFile(req, upload)
	if err != nil {
		app.Log.Errorf("Assembly failed for request %s, couldn't create file in database: %s", req.FlowIdentifier, err)
		return nil, err
	}

	// Create on disk entry to write chunks to
	dest, err := s.createDest(file.ID)
	if err != nil {
		app.Log.Errorf("Assembly failed for request %s, couldn't create file on disk: %s", req.FlowIdentifier, err)
		return file, err
	}

	// Assemble the chunks
	chunkSupplier := newDirChunkSupplier(dir)
	if err := assembleRequest(chunkSupplier, dest); err != nil {
		app.Log.Errorf("Assembly failed for request %s, couldn't assemble request: %s", req.FlowIdentifier, err)
		return file, err
	}

	// Finish updating the file state.
	finisher := newFinisher(s.files, s.dirs)
	if err := finisher.finish(req, file.ID, upload); err != nil {
		app.Log.Errorf("Assembly failed for request %s, couldn't finish request: %s", req.FlowIdentifier, err)
		return file, err
	}

	app.Log.Infof("successfully upload fileID %s", file.ID)

	s.cleanupUploadRequest(req.UploadID())
	return nil, nil
}

// createFile creates the database file entry.
func (s *uploadService) createFile(req *UploadRequest, upload *schema.Upload) (*schema.File, error) {
	file := schema.NewFile(upload.File.Name, upload.ProjectOwner)

	f, err := s.files.Insert(&file, upload.DirectoryID, upload.ProjectID)
	app.Log.Infof("Created file %s, in %s %s", f.ID, upload.DirectoryID, upload.ProjectID)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// createDest creates the destination file and ensures that the directory
// path is also created.
func (s *uploadService) createDest(fileID string) (io.Writer, error) {
	dir := app.MCDir.FileDir(fileID)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	return os.Create(app.MCDir.FilePath(fileID))
}

// requestDir returns the directory to write this requests chunks to.
func (s *uploadService) requestDir(req *flow.Request) string {
	requestPath := &mcdirRequestPath{}
	return requestPath.Dir(req)
}

// Write will write a chunk to the given directory.
func (s *uploadService) Write(dest string, req *flow.Request) error {
	writer := &fileRequestWriter{}
	return writer.Write(dest, req)
}

// cleanup is called when an error has occurred. It attempts to clean up
// the state in the database for this particular entry.
func (s *uploadService) cleanup(req *UploadRequest, fileID string) error {
	upload, err := s.uploads.ByID(req.FlowIdentifier)
	if err != nil {
		return err
	}
	_, err = s.files.Delete(fileID, upload.DirectoryID, upload.ProjectID)
	return err
}

//cleanupUploadRequest removes the upload request and file chunks.
func (s *uploadService) cleanupUploadRequest(uploadID string) {

	s.tracker.clear(uploadID)
	s.uploads.Delete(uploadID)
	os.RemoveAll(app.MCDir.UploadDir(uploadID))
}
