package mc

import (
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/server/mcstore"
)

type ClientAPI struct {
	serverAPI     *mcstore.ServerAPI
	projectOpener ProjectDBOpener
}

func NewClientAPI() *ClientAPI {
	return &ClientAPI{
		serverAPI:     mcstore.NewServerAPI(),
		projectOpener: ProjectOpener,
	}
}

func newClientAPIWithConfiger(configer Configer) *ClientAPI {
	opener := sqlProjectDBOpener{
		configer: configer,
	}
	return &ClientAPI{
		serverAPI:     mcstore.NewServerAPI(),
		projectOpener: opener,
	}
}

func (c *ClientAPI) UploadFile(projectID string, path string) error {
	return nil
}

func (c *ClientAPI) UploadDirectory(projectID string, path string) error {
	return nil
}

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

func (c *ClientAPI) ProjectStatus(projectID string) error {
	return nil
}

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

func (c *ClientAPI) CreateDirectory(projectName, path string) (string, error) {
	projectDB, err := c.projectOpener.OpenProjectDB(projectName)
	if err != nil {
		return "", err
	}

	return c.createDir(projectDB.Project(), path)
}

func (c *ClientAPI) createDir(project *Project, path string) (directoryID string, err error) {
	req := mcstore.DirectoryRequest{
		ProjectName: project.Name,
		ProjectID:   project.ProjectID,
		Path:        path,
	}

	return c.serverAPI.GetDirectory(req)
}

func (c *ClientAPI) RenameProject(oldName, newName string) error {
	return nil
}

func (c *ClientAPI) IndexProject(projectName string) error {
	return nil
}
