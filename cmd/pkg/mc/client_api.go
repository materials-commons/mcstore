package mc

import (
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
)

// ClientAPI implements API calls to the mcstored server.
type ClientAPI struct {
	serverAPI     *mcstoreapi.ServerAPI
	projectOpener ProjectDBOpener
}

// NewClientAPI creates a new instance of a ClientAPI. It checks for client side project
// data in $HOME/.materialscommons
func NewClientAPI() *ClientAPI {
	return &ClientAPI{
		serverAPI:     mcstoreapi.NewServerAPI(),
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
		serverAPI:     mcstoreapi.NewServerAPI(),
		projectOpener: opener,
	}
}

// UploadFile uploads a single file to the given project.
func (c *ClientAPI) UploadFile(projectName string, path string) error {
	if projectDB, err := ProjectOpener.OpenProjectDB(projectName); err != nil {
		return err
	} else {
		uploader := &projectUploader{
			db:         projectDB,
			numThreads: 1,
		}
		return uploader.uploadFile(path)
	}
}

// UploadDirectory uploads all the entries in a given directory. It will
// not follow sub directories. However it will create sub directories
// that are direct children of the given path.
func (c *ClientAPI) UploadDirectory(projectName string, path string, recursive bool, numThreads int) error {
	if projectDB, err := ProjectOpener.OpenProjectDB(projectName); err != nil {
		return err
	} else {
		uploader := &projectUploader{
			db:         projectDB,
			numThreads: numThreads,
		}
		return uploader.uploadDirectory(path, recursive)
	}
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
		return uploader.uploadProject()
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

	req := mcstoreapi.CreateProjectRequest{
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
		if err == nil && finfo.IsDir() {
			c.createDirectory(projectDB, path)
		}
		return nil
	})

	return nil
}

// CreateDirectory will create a single directory on the server for the named project.
func (c *ClientAPI) CreateDirectory(projectName, path string) error {
	projectDB, err := c.projectOpener.OpenProjectDB(projectName)
	if err != nil {
		return err
	}
	return c.createDirectory(projectDB, path)
}

func (c *ClientAPI) createDirectory(db ProjectDB, path string) error {
	if dirID, err := c.getDir(db.Project(), path); err != nil {
		return err
	} else {
		if _, err := db.FindDirectory(path); err == app.ErrNotFound {
			dir := &Directory{
				DirectoryID: dirID,
				Path:        path,
			}
			db.InsertDirectory(dir)
		}
		return nil
	}
}

// createDir performs the actual call to the server to create a directory.
func (c *ClientAPI) getDir(project *Project, path string) (directoryID string, err error) {
	req := mcstoreapi.DirectoryRequest{
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
