package uploads

import (
	"io"

	"github.com/materials-commons/gohandy/file"
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
	tracker     tracker
	files       dai.Files
	uploads     dai.Uploads
	dirs        dai.Dirs
	writer      requestWriter
	requestPath requestPath
	fops        file.Operations
}

// NewUploadService creates a new uploadService connecting
// to the default database. It will panic if it cannot
// establish a connection to the database.
func NewUploadService() *uploadService {
	session := db.RSessionMust()
	return &uploadService{
		tracker:     requestBlockCountTracker,
		files:       dai.NewRFiles(session),
		uploads:     dai.NewRUploads(session),
		dirs:        dai.NewRDirs(session),
		writer:      &fileRequestWriter{},
		requestPath: &mcdirRequestPath{},
		fops:        file.OS,
	}
}

// Upload performs uploading a block and constructing the file
// after all blocks have been uploaded.
func (s *uploadService) Upload(req *UploadRequest) error {
	dir := s.requestPath.dir(req.Request)
	if err := s.writer.write(dir, req.Request); err != nil {
		return err
	}

	id := req.UploadID()
	s.tracker.setBlock(id, int(req.FlowChunkNumber))

	if s.tracker.done(id) {
		// assembling the file, and performing any processing may take a while, especially
		// on large uploads, so we perform this step in the background.
		go func(req *UploadRequest, dir string) {
			if file, err := s.assemble(req, dir); err != nil {
				app.Log.Errorf("Assembly failed for request %s: %s", req.FlowIdentifier, err)
				// Assembly failed. If file isn't nil then
				// there is some cleanup to do in the database.
				if file != nil {
					if err := s.cleanup(req, file.ID); err != nil {
						app.Log.Errorf("Attempted cleanup of failed assembly %s errored with: %s", req.FlowIdentifier, err)
					}
				}
			}
		}(req, dir)
	}
	return nil
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
	if err := s.fops.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	return s.fops.Create(app.MCDir.FilePath(fileID))
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
	s.fops.RemoveAll(app.MCDir.UploadDir(uploadID))
}
