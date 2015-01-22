package dai

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Users interface {
	ByAPIKey(apikey string) (*schema.User, error)
}

type Files interface {
	ByID(id string) (*schema.File, error)
	ByChecksum(checksum string) (*schema.File, error)
	Insert(file *schema.File, dirID string, projectID string) (*schema.File, error)
	Update(file *schema.File) error
	UpdateFields(fileID string, fields map[string]interface{}) error
}

type Groups interface {
	ByID(id string) (*schema.Group, error)
	ForOwner(owner string) ([]schema.Group, error)
}

type Uploads interface {
	ByID(id string) (*schema.Upload, error)
	Insert(upload *schema.Upload) (*schema.Upload, error)
	Update(upload *schema.Upload) error
	ForUser(user string) ([]schema.Upload, error)
	Delete(uploadID string) error
}

type Projects interface {
	ByID(id string) (*schema.Project, error)
	HasDirectory(projectID, directoryID string) bool
}

type Dirs interface {
	ByID(id string) (*schema.Directory, error)
}
