package uploads

import (
	"os"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type UploadRequest struct {
	*flow.Request
	Owner string
}

type UploadService interface {
	Upload(req *UploadRequest) error
}

type uploadService struct {
	tracker *uploadTracker
	files   dai.Files
	uploads dai.Uploads
}

func NewUploadService() *uploadService {
	return &uploadService{}
}

func (s *uploadService) Upload(req *UploadRequest) error {
	dir := s.requestDir(req.Request)

	if s.allBlocksUploaded(req.UploadID(), req.FlowTotalChunks) {
		s.assemble(req, dir)
		return nil
	}

	if err := s.Write(dir, req.Request); err != nil {
		return err
	}

	id := req.UploadID()
	s.tracker.increment(id)
	return nil
}

func (s *uploadService) allBlocksUploaded(id string, totalChunks int32) bool {
	count := s.tracker.count(id)
	return count == totalChunks
}

func (s *uploadService) assemble(req *UploadRequest, dir string) {
	file, err := s.createFile(req)
	if err != nil {
		// log
		return
	}

	dest, err := os.Create(app.MCDir.FilePath(file.ID))
	if err != nil {
		// log
		return
	}

	chunkSupplier := newDirChunkSupplier(dir)
	if err := assembleRequest(chunkSupplier, dest); err != nil {
		// log
		return
	}

	finisher := newFinisher(nil)
	if err := finisher.finish(req, file.ID, req.DirectoryID); err != nil {
		// log
		return
	}

	s.tracker.clear(req.UploadID())
}

func (s *uploadService) createFile(req *UploadRequest) (*schema.File, error) {
	upload, err := s.uploads.ByID(req.FlowIdentifier)
	if err != nil {
		return nil, err
	}
	file := schema.NewFile(req.FlowFileName, req.Owner)

	f, err := s.files.Insert(&file, upload.DirectoryID, upload.ProjectID)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *uploadService) requestDir(req *flow.Request) string {
	requestPath := &mcdirRequestPath{}
	return requestPath.Dir(req)
}

func (s *uploadService) Write(dest string, req *flow.Request) error {
	writer := &fileRequestWriter{}
	return writer.Write(dest, req)
}
