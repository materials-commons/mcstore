package uploads

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
)

type UploadsService interface {
	Create(req CreateRequest) (*schema.Upload, error)
}

type uploadsService struct {
	dirs     dai.Dirs
	projects dai.Projects
	uploads  dai.Uploads
	access   domain.Access
}

func NewUploadsService() *uploadsService {
	session := db.RSessionMust()
	access := domain.NewAccess(dai.NewRGroups(session), dai.NewRFiles(session), dai.NewRUsers(session))
	return &uploadsService{
		dirs:     dai.NewRDirs(session),
		projects: dai.NewRProjects(session),
		uploads:  dai.NewRUploads(session),
		access:   access,
	}
}

func NewUploadsServiceFrom(dirs dai.Dirs, projects dai.Projects, uploads dai.Uploads, access domain.Access) *uploadsService {
	return &uploadsService{
		dirs:     dirs,
		projects: projects,
		uploads:  uploads,
		access:   access,
	}
}

func (s *uploadsService) Create(req CreateRequest) (*schema.Upload, error) {
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
		Create()
	return s.uploads.Insert(&upload)
}

func (s *uploadsService) getProj(projectID, user string) (*schema.Project, error) {
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

func (s *uploadsService) getDir(directoryID, projectID, user string) (*schema.Directory, error) {
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
