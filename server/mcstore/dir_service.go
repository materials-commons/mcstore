package mcstore

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
)

// DirService creates or retrieves directories in a project
type DirService interface {
	createDir(projectID, path string) (*schema.Directory, error)
}

// dirService implements the DirService interface
type dirService struct {
	dirs     dai.Dirs
	projects dai.Projects
	access   domain.Access
}

// NewDirService creates a new dirService. It uses db.RSessionMust() to
// create a session for the database. It will panic if it cannot connect
// to the database.
func newDirService() *dirService {
	session := db.RSessionMust()

	access := domain.NewAccess(dai.NewRProjects(session), dai.NewRFiles(session), dai.NewRUsers(session))
	return &dirService{
		dirs:     dai.NewRDirs(session),
		projects: dai.NewRProjects(session),
		access:   access,
	}
}

// CreateDir will look up a given directory path for a project. If that
// path exists it will return the directory. If the path doesn't exist
// then it will create the directory and return it. CreateDir validates
// the path and returns an error if the path is not valid for the project.
func (s *dirService) createDir(projectID, path string) (*schema.Directory, error) {
	proj, err := s.projects.ByID(projectID)
	if err != nil {
		return nil, err
	} else if !validDirPath(proj.Name, path) {
		return nil, app.ErrInvalid
	}

	dir, err := s.dirs.ByPath(path, projectID)
	switch {
	case err == app.ErrNotFound:
		// Doesn't exist, create it
		parent := filepath.Dir(path)
		d := schema.NewDirectory(path, proj.Owner, projectID, parent)
		dir, err = s.dirs.Insert(&d)
		return dir, err
	case err != nil:
		return nil, err
	default:
		// Existing directory found, so just return it.
		return dir, nil
	}
}

// validDirPath verifies that the directory path starts with the project name.
// It handles both Linux (/) and Windows (\) style slashes.
func validDirPath(projName, dirPath string) bool {
	slash := strings.Index(dirPath, "/")
	if slash == -1 {
		slash = strings.Index(dirPath, "\\")
	}
	switch {
	case slash == -1:
		return false
	case projName != dirPath[:slash]:
		return false
	default:
		return true
	}
}
