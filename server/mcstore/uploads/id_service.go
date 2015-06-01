package uploads

import (
	"time"

	"math"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/willf/bitset"
)

// A IDRequest requests a new upload id be created for
// the given parameters.
type IDRequest struct {
	User        string
	DirectoryID string
	ProjectID   string
	FileName    string
	FileSize    int64
	Checksum    string
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
	access      domain.Access
	fops        file.Operations
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
		access:      access,
		fops:        file.OS,
		requestPath: &mcdirRequestPath{},
	}
}

// ID will create a new Upload request. It validates and checks access to the given project
// and directory.
func (s *idService) ID(req IDRequest) (*schema.Upload, error) {
	proj, err := s.getProj(req.ProjectID, req.User)
	if err != nil {
		return nil, err
	}

	dir, err := s.getDir(req.DirectoryID, proj.ID, req.User)
	if err != nil {
		return nil, err
	}

	searchParams := dai.UploadSearch{
		ProjectID:   proj.ID,
		DirectoryID: dir.ID,
		FileName:    req.FileName,
		Checksum:    req.Checksum,
	}

	if existingUpload, err := s.uploads.Search(searchParams); err != nil {
		return existingUpload, nil
	}

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
		Create()
	u, err := s.uploads.Insert(&upload)
	if err != nil {
		return nil, err
	}

	if err := s.initUpload(req, u.ID); err != nil {
		s.uploads.Delete(u.ID)
		return nil, err
	}

	return u, nil
}

// getProj retrieves the project with the given projectID. It checks that the
// given user has access to that project.
func (s *idService) getProj(projectID, user string) (*schema.Project, error) {
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
func (s *idService) getDir(directoryID, projectID, user string) (*schema.Directory, error) {
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
func (s *idService) initUpload(req IDRequest, id string) error {
	if err := s.requestPath.mkdirFromID(id); err != nil {
		return err
	}

	bset := bitset.New(numBlocks(req.FileSize))
	f, err := s.fops.Create(BlocksFile(s.requestPath, id))
	if err != nil {
		return err
	}
	defer f.Close()
	bset.WriteTo(f)
	return nil
}

// TODO: Fix assumption of twoMeg chunks. This is used in a few places in code.
const twoMeg = 2 * 1024 * 1024

// numBlocks
func numBlocks(fileSize int64) uint {
	// round up to nearest number of blocks
	d := float64(fileSize) / float64(twoMeg)
	return uint(math.Ceil(d))
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
	_, err := s.getProj(projectID, user)
	if err != nil {
		return nil, err
	}
	return s.uploads.ForProject(projectID)
}
