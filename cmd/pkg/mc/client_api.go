package mc

import (
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/server/mcstore"
)

// ClientAPI implements API calls to the mcstored server.
type ClientAPI struct {
	serverAPI     *mcstore.ServerAPI
	projectOpener ProjectDBOpener
}

// NewClientAPI creates a new instance of a ClientAPI. It checks for client side project
// data in $HOME/.materialscommons
func NewClientAPI() *ClientAPI {
	return &ClientAPI{
		serverAPI:     mcstore.NewServerAPI(),
		projectOpener: ProjectOpener,
	}
}

// newClientAPIWithConfiger creates a new instance of a ClientAPI. The calls supplies the
// configer used to find the client side project database.
func newClientAPIWithConfiger(configer Configer) *ClientAPI {
	opener := sqlProjectDBOpener{
		configer: configer,
	}
	return &ClientAPI{
		serverAPI:     mcstore.NewServerAPI(),
		projectOpener: opener,
	}
}

// UploadFile uploads a single file to the given project.
func (c *ClientAPI) UploadFile(projectName string, path string) error {
	return nil
}

// UploadDirectory uploads all the entries in a given directory. It will
// not follow sub directories. However it will create sub directories
// that are direct children of the given path.
func (c *ClientAPI) UploadDirectory(projectName string, path string) error {
	return nil
}

// UploadProject will upload all the changed and new files in a given project.
func (c *ClientAPI) UploadProject(projectName string, numThreads int) error {
	if projectDB, err := ProjectOpener.OpenProjectDB(projectName); err != nil {
		return err
	} else {
		uploader := &projectUploader{
			db:         projectDB,
			numThreads: numThreads,
		}
		return uploader.upload()
	}
}

//
func (c *ClientAPI) ProjectStatus(projectID string) error {
	return nil
}

// CreateProject creates a new project. If the project already
// exists on the client it returns app.ErrExists.
func (c *ClientAPI) CreateProject(name, path string) error {
	if c.projectOpener.ProjectExists(name) {
		return app.ErrExists
	}

	req := mcstore.CreateProjectRequest{
		Name: name,
	}
	if resp, err := c.serverAPI.CreateProject(req); err != nil {
		return err
	} else {
		projectDBSpec := ProjectDBSpec{
			Name:      name,
			ProjectID: resp.ProjectID,
			Path:      path,
		}
		_, err := c.projectOpener.CreateProjectDB(projectDBSpec)
		return err
	}
}

// CreateProjectDirectories will create all the directories on the server
// that are found under the client project.
func (c *ClientAPI) CreateProjectDirectories(projectName string) error {
	projectDB, err := ProjectOpener.OpenProjectDB(projectName)
	if err != nil {
		return err
	}

	project := projectDB.Project()
	filepath.Walk(project.Path, func(path string, finfo os.FileInfo, err error) error {
		if err != nil && finfo.IsDir() {
			c.createDir(project, path)
		}
		return nil
	})

	return nil
}

// CreateDirectory will create a single directory on the server for the named project.
func (c *ClientAPI) CreateDirectory(projectName, path string) (string, error) {
	projectDB, err := c.projectOpener.OpenProjectDB(projectName)
	if err != nil {
		return "", err
	}

	return c.createDir(projectDB.Project(), path)
}

// createDir performs the actual call to the server to create a directory.
func (c *ClientAPI) createDir(project *Project, path string) (directoryID string, err error) {
	req := mcstore.DirectoryRequest{
		ProjectName: project.Name,
		ProjectID:   project.ProjectID,
		Path:        path,
	}

	return c.serverAPI.GetDirectory(req)
}

// RenameProject will rename an existing project.
func (c *ClientAPI) RenameProject(oldName, newName string) error {
	return nil
}

// IndexProject will index a project (whatever that means at the moment)
func (c *ClientAPI) IndexProject(projectName string) error {
	return nil
}
