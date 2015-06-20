package uploads

import (
	"time"

	"math"

	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
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
	tracker     tracker
	requestPath requestPath
}

// NewIDService creates a new idService. It uses db.RSessionMust() to get
// a session to connect to the database. It will panic if it cannot connect to
// the database.
func NewIDService() *idService {
	session := db.RSessionMust()
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

// NewIDServiceUsingSession creates a new idService that connects to the database using
// the given session.
func NewIDServiceUsingSession(session *r.Session) *idService {
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

func (s *idService) ID2(req IDRequest) (*schema.Upload, error) {
	var (
		err  error
		proj *schema.Project
		dir  *schema.Directory
	)

	if proj, err = s.getProjectValidatingAccess(req.ProjectID, req.User); err != nil {
		return nil, err
	}

	if dir, err = s.getDirectoryValidatingAccess(req.DirectoryID, req.ProjectID, req.User); err != nil {
		return nil, err
	}

	upload, err := s.findExisting(req, proj, dir)
	switch {
	case err == app.ErrNotFound:
		return s.createNewUploadRequest(req, proj, dir)
	case err != nil:
		return nil, err
	default:
		return upload, nil
	}
}

func (s *idService) findExisting(req IDRequest, proj *schema.Project, dir *schema.Directory) (*schema.Upload, error) {
	if file, err := s.files.ByChecksum(req); err == nil {
		return s.createUploadFromFile(req, file)
	}
	return s.findMatchingUploadRequest(req)
}

func (s *idService) createUploadFromFile(req IDRequest, file *schema.File) (*schema.Upload, error) {
	// We have the real uploaded file. There are 4 cases:
	//   1. This file matches the project, directory and name we want to upload, so
	//      we create a finished upload request because there is nothing to do.
	//
	//   2. This file matches the project and directory, but not name, so we create
	//      a new file entry pointing at this one and return a finished upload request.
	//
	//   3. This entry is in a different project and/or directory and there isn't a matching
	//      one in the given project/directory, so create a new file entry pointing at this
	//      one and return a finished upload request.
	//
	//   4. This entry is in a different project and/or directory and there is a matching
	//      one in the given project/directory. Return a finished upload request.
	return nil, nil
}

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

func (s *idService) createNewUploadRequest(req IDRequest, proj *schema.Project, dir *schema.Directory) (*schema.Upload, error) {
	n := uint(numBlocks(req.FileSize, req.ChunkSize))
	upload := schema.CUpload().
		Owner(req.User).
		Project(req.ProjectID, proj.Name).
		ProjectOwner(proj.Owner).
		Directory(req.DirectoryID, dir.Name).
		Host(req.Host).
		FName(req.FileName).
		FSize(req.FileSize).
		FChecksum(req.Checksum).
		FRemoteMTime(req.FileMTime).
		FBlocks(bitset.New(n)).
		Create()
	u, err := s.uploads.Insert(&upload)
	if err != nil {
		fmt.Println("insert err", err)
		return nil, err
	}

	if err := s.initUpload(u.ID, req.FileSize, req.ChunkSize); err != nil {
		fmt.Println("initUpload err", err)
		s.uploads.Delete(u.ID)
		return nil, err
	}
	return u, nil
}

// ID will create a new Upload request or return an existing one.
func (s *idService) ID(req IDRequest) (*schema.Upload, error) {
	///files, err := s.files.AllByChecksum(req.Checksum)
	proj, err := s.getProjectValidatingAccess(req.ProjectID, req.User)
	if err != nil {
		fmt.Println("getProj err", err)
		return nil, err
	}

	dir, err := s.getDirectoryValidatingAccess(req.DirectoryID, proj.ID, req.User)
	if err != nil {
		fmt.Println("getDir err", err)
		return nil, err
	}

	searchParams := dai.UploadSearch{
		ProjectID:   proj.ID,
		DirectoryID: dir.ID,
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

	n := uint(numBlocks(req.FileSize, req.ChunkSize))
	upload := schema.CUpload().
		Owner(req.User).
		Project(req.ProjectID, proj.Name).
		ProjectOwner(proj.Owner).
		Directory(req.DirectoryID, dir.Name).
		Host(req.Host).
		FName(req.FileName).
		FSize(req.FileSize).
		FChecksum(req.Checksum).
		FRemoteMTime(req.FileMTime).
		FBlocks(bitset.New(n)).
		Create()
	u, err := s.uploads.Insert(&upload)
	if err != nil {
		fmt.Println("insert err", err)
		return nil, err
	}

	if err := s.initUpload(u.ID, req.FileSize, req.ChunkSize); err != nil {
		fmt.Println("initUpload err", err)
		s.uploads.Delete(u.ID)
		return nil, err
	}
	return u, nil
}

func (s *idService) getAndValidateDependencies(req IDRequest) (*schema.Project, *schema.Directory, error) {
	proj, err := s.getProjectValidatingAccess(req.ProjectID, req.User)
	if err != nil {
		fmt.Println("getProj err", err)
		return nil, nil, err
	}

	dir, err := s.getDirectoryValidatingAccess(req.DirectoryID, proj.ID, req.User)
	if err != nil {
		fmt.Println("getDir err", err)
		return nil, nil, err
	}

	return proj, dir, nil
}

// getProj retrieves the project with the given projectID. It checks that the
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

// getDir retrieves the directory with the given directoryID. It checks access to the
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

// initUpload
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
	// getProj will validate the project and access.
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
