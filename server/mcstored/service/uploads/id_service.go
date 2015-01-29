package uploads

import (
	"time"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
)

// A IDRequest requests a new upload id be created for
// the given parameters.
type IDRequest struct {
	User        string
	DirectoryID string
	ProjectID   string
	FileName    string
	FileSize    int64
	FileMTime   time.Time
	Host        string
	Birthtime   time.Time
}

// IDService creates new upload requests
type IDService interface {
	ID(req IDRequest) (*schema.Upload, error)
}

// idService implements the IDService interface using
// the dai services.
type idService struct {
	dirs     dai.Dirs
	projects dai.Projects
	uploads  dai.Uploads
	access   domain.Access
}

// NewIDService creates a new idService. It uses db.RSessionMust() to get
// a session to connect to the database. It will panic if it cannot connect to
// the database.
func NewIDService() *idService {
	session := db.RSessionMust()
	access := domain.NewAccess(dai.NewRGroups(session), dai.NewRFiles(session), dai.NewRUsers(session))
	return &idService{
		dirs:     dai.NewRDirs(session),
		projects: dai.NewRProjects(session),
		uploads:  dai.NewRUploads(session),
		access:   access,
	}
}

// NewIDServiceFrom creates a new instance of the idService using the passed in dai and access parameters.
func NewIDServiceFrom(dirs dai.Dirs, projects dai.Projects, uploads dai.Uploads, access domain.Access) *idService {
	return &idService{
		dirs:     dirs,
		projects: projects,
		uploads:  uploads,
		access:   access,
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

	upload := schema.CUpload().
		Owner(req.User).
		Project(req.ProjectID, proj.Name).
		Directory(req.DirectoryID, dir.Name).
		Host(req.Host).
		FName(req.FileName).
		FSize(req.FileSize).
		FRemoteMTime(req.FileMTime).
		Create()
	return s.uploads.Insert(&upload)
}

// getProj retrieves the project with the given projectID. It checks that the
// given user has access to that project.
func (s *idService) getProj(projectID, user string) (*schema.Project, error) {
	project, err := s.projects.ByID(projectID)
	switch {
	case err != nil:
		return nil, err
	case !s.access.AllowedByOwner(project.Owner, user):
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
	case !s.access.AllowedByOwner(dir.Owner, user):
		return nil, app.ErrNoAccess
	case !s.projects.HasDirectory(projectID, directoryID):
		return nil, app.ErrInvalid
	default:
		return dir, nil
	}
}