package mc

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/server/mcstore"
)

type ClientAPI struct {
	serverAPI *mcstore.ServerAPI
}

func NewClientAPI() *ClientAPI {
	return &ClientAPI{
		serverAPI: mcstore.NewServerAPI(),
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
	if ProjectOpener.ProjectExists(name) {
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
		_, err := ProjectOpener.CreateProjectDB(projectDBSpec)
		return err
	}
}

func (c *ClientAPI) CreateDirectory(projectName, path string) error {
	return nil
}

func (c *ClientAPI) RenameProject(oldName, newName string) error {
	return nil
}

func (c *ClientAPI) IndexProject(projectName string) error {
	return nil
}
