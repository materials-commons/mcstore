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

func (c *ClientAPI) UploadProject(projectID string) error {
	return nil
}

func (c *ClientAPI) ProjectStatus(projectID string) error {
	return nil
}

type Project struct {
	Name      string
	Path      string
	ProjectID string
}

func (c *ClientAPI) CreateProject(project Project) error {
	return nil
}

type Directory struct {
	Path        string
	ProjectName string
}

func (c *ClientAPI) CreateDirectory(dir Directory) error {
	return nil
}

func (c *ClientAPI) RenameProject(oldName, newName string) error {
	return nil
}

func (c *ClientAPI) IndexProject(projectName string) error {
	return nil
}
