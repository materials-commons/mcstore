package uploads

import (
	"time"

	"math"

	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/willf/bitset"
)

var _ = fmt.Println

// A IDRequest requests a new upload id be created for
// the given parameters.
type IDRequest struct {
	User        string
	DirectoryID string
	ProjectID   string
	FileName    string
	FileSize    int64
	Checksum    string
	ChunkSize   int32
	FileMTime   time.Time
	Host        string
	Birthtime   time.Time
}

// IDService creates new upload requests
type IDService interface {
	ID(req IDRequest) (*schema.Upload, error)
	Delete(requestID, user string) error
	UploadsForProject(projectID, user string) ([]schema.Upload, error)
}

// idService implements the IDService interface using
// the dai services.
type idService struct {
	dirs        dai.Dirs
	projects    dai.Projects
	uploads     dai.Uploads
	files       dai.Files
	access      domain.Access
	fops        file.Operations
	tracker     *blockTracker
	requestPath requestPath
}

// NewIDService creates a new idService that connects to the database using
// the given session.
func NewIDService(session *r.Session) *idService {
	access := domain.NewAccess(dai.NewRProjects(session), dai.NewRFiles(session), dai.NewRUsers(session))
	return &idService{
		dirs:        dai.NewRDirs(session),
		projects:    dai.NewRProjects(session),
		uploads:     dai.NewRUploads(session),
		files:       dai.NewRFiles(session),
		access:      access,
		fops:        file.OS,
		tracker:     requestBlockTracker,
		requestPath: &mcdirRequestPath{},
	}
}

// ID will create a new Upload request or return an existing one.
func (s *idService) ID(req IDRequest) (*schema.Upload, error) {
	var (
		err  error
		proj *schema.Project
		dir  *schema.Directory
	)

	// Check that project exists and user has access
	if proj, err = s.getProjectValidatingAccess(req.ProjectID, req.User); err != nil {
		return nil, err
	}

	// Check that directory exists and user has access
	if dir, err = s.getDirectoryValidatingAccess(req.DirectoryID, req.ProjectID, req.User); err != nil {
		return nil, err
	}

	return s.createUploadRequest(req, proj, dir)
}

// createUploadRequest will create an upload. It checks if there is a matching file or an existing upload.
// If it finds a matching file it returns a finished upload. If it finds an existing upload it returns it,
// otherwise it returns a new upload.
func (s *idService) createUploadRequest(req IDRequest, proj *schema.Project, dir *schema.Directory) (*schema.Upload, error) {
	upload, err := s.findExisting(req, proj, dir)
	switch {
	case err == app.ErrNotFound:
		upload := s.prepareUploadRequest(req, proj, dir)
		return s.finishUploadRequest(upload)
	case err != nil:
		return nil, err
	default:
		return upload, nil
	}
}

// findExisting checks if there is an outstanding upload request, or a file that already matches
// the upload request. If it finds a match it will return the corresponding upload request. In
// the case of an already uploaded file it will return an upload request with all blocks marked
// as uploaded.
func (s *idService) findExisting(req IDRequest, proj *schema.Project, dir *schema.Directory) (*schema.Upload, error) {
	if _, err := s.files.ByChecksum(req.Checksum); err == nil {
		return s.createFinishedUpload(req, proj, dir)
	}
	return s.findMatchingUploadRequest(req)
}

// createFinishedUpload will create an upload entry with all blocks marked as uploaded.
func (s *idService) createFinishedUpload(req IDRequest, proj *schema.Project, dir *schema.Directory) (*schema.Upload, error) {
	// Create a new upload request and then set all blocks as already uploaded.
	upload := s.prepareUploadRequest(req, proj, dir)
	upload.IsExisting = true
	s.tracker.markAllBlocks(upload.ID)
	s.tracker.setIsExistingFile(upload.ID, true)
	upload.SetFBlocks(s.tracker.getBlocks(upload.ID))
	return s.finishUploadRequest(upload)
}

// findMatchingUploadRequest will search the set of upload requests and see if there
// is already an outstanding upload request that matches this one.
func (s *idService) findMatchingUploadRequest(req IDRequest) (*schema.Upload, error) {
	searchParams := dai.UploadSearch{
		ProjectID:   req.ProjectID,
		DirectoryID: req.DirectoryID,
		FileName:    req.FileName,
		Checksum:    req.Checksum,
	}

	if existingUpload, err := s.uploads.Search(searchParams); err == nil {
		// Found existing
		if bset := s.tracker.getBlocks(existingUpload.ID); bset != nil {
			existingUpload.File.Blocks = bset
		}
		return existingUpload, nil
	}
	return nil, app.ErrNotFound
}

// prepareUploadRequest creates a new upload request.
func (s *idService) prepareUploadRequest(req IDRequest, proj *schema.Project, dir *schema.Directory) *schema.Upload {
	n := uint(numBlocks(req.FileSize, req.ChunkSize))
	upload := schema.CUpload().
		Owner(req.User).
		Project(req.ProjectID, proj.Name).
		ProjectOwner(proj.Owner).
		Directory(req.DirectoryID, dir.Name).
		Host(req.Host).
		FName(req.FileName).
		FSize(req.FileSize).
		FChunk(int(req.ChunkSize), int(n)).
		FChecksum(req.Checksum).
		FRemoteMTime(req.FileMTime).
		FBlocks(bitset.New(n)).
		Create()
	return &upload
}

// finishUploadRequest takes a prepared upload request, inserts it into the database and
// initializes the state associated with the upload. If the upload fails to initialize
// it will delete the upload request from the database.
func (s *idService) finishUploadRequest(upload *schema.Upload) (*schema.Upload, error) {
	u, err := s.uploads.Insert(upload)
	if err != nil {
		return nil, err
	}

	if err := s.initUpload(u.ID, u.File.Size, int32(u.File.ChunkSize)); err != nil {
		s.uploads.Delete(u.ID)
		return nil, err
	}
	return u, nil
}

// getProjectValidatingAccess retrieves the project with the given projectID. It checks that the
// given user has access to that project.
func (s *idService) getProjectValidatingAccess(projectID, user string) (*schema.Project, error) {
	project, err := s.projects.ByID(projectID)
	switch {
	case err != nil:
		return nil, err
	case !s.access.AllowedByOwner(projectID, user):
		return nil, app.ErrNoAccess
	default:
		return project, nil
	}
}

// getDirectoryValidatingAccess retrieves the directory with the given directoryID. It checks access to the
// directory and validates that the directory exists in the given project.
func (s *idService) getDirectoryValidatingAccess(directoryID, projectID, user string) (*schema.Directory, error) {
	dir, err := s.dirs.ByID(directoryID)
	switch {
	case err != nil:
		return nil, err
	case !s.projects.HasDirectory(projectID, directoryID):
		return nil, app.ErrInvalid
	case !s.access.AllowedByOwner(projectID, user):
		return nil, app.ErrNoAccess
	default:
		return dir, nil
	}
}

// initUpload initializes the upload state. It creates the directory to write
// the upload blocks to and creates a tracker entry for the upload.
func (s *idService) initUpload(id string, fileSize int64, chunkSize int32) error {
	if err := s.requestPath.mkdirFromID(id); err != nil {
		return err
	}

	s.tracker.load(id, numBlocks(fileSize, chunkSize))
	return nil
}

// numBlocks
func numBlocks(fileSize int64, chunkSize int32) int {
	// round up to nearest number of blocks
	d := float64(fileSize) / float64(chunkSize)
	n := int(math.Ceil(d))
	return n
}

// Delete will delete the given requestID if the user has access
// to delete that request. Owners of requests can delete their
// own requests. Project owners can delete any request, even if
// they don't own the request.
func (s *idService) Delete(requestID, user string) error {
	upload, err := s.uploads.ByID(requestID)
	switch {
	case err != nil:
		return err
	case !s.access.AllowedByOwner(upload.ProjectID, user):
		return app.ErrNoAccess
	default:
		if err := s.uploads.Delete(requestID); err != nil {
			return err
		}
		// Delete the directory where chunks were being written
		s.fops.RemoveAll(app.MCDir.UploadDir(requestID))
		return nil
	}
}

// ListForProject will return all the uploads associated with a project.
func (s *idService) UploadsForProject(projectID, user string) ([]schema.Upload, error) {
	_, err := s.getProjectValidatingAccess(projectID, user)
	switch {
	case err == app.ErrNotFound:
		// Invalid project
		return nil, app.ErrInvalid
	case err != nil:
		return nil, err
	default:
		return s.uploads.ForProject(projectID)
	}
}
