package mc

import "github.com/materials-commons/mcstore/server/mcstore"

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
	projectDB, err := ProjectOpener.OpenProjectDB(projectName)
	if err != nil {
		return err
	}
	uploader := &projectUploader{
		db:         projectDB,
		numThreads: numThreads,
	}
	return uploader.upload()
}

func (c *ClientAPI) ProjectStatus(projectID string) error {
	return nil
}

func (c *ClientAPI) CreateProject(projectSpec ProjectDBSpec) error {
	_, err := ProjectOpener.CreateProjectDB(projectSpec)
	return err
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
