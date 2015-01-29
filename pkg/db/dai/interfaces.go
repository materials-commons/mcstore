package dai

import "github.com/materials-commons/mcstore/pkg/db/schema"

// Users gives access to users.
type Users interface {
	ByAPIKey(apikey string) (*schema.User, error)
}

// Files allows manipulation and access to file.
type Files interface {
	ByID(id string) (*schema.File, error)
	ByChecksum(checksum string) (*schema.File, error)
	ByPath(name, dirID string) (*schema.File, error)
	Directories(fileID string) ([]string, error)
	Insert(file *schema.File, dirID string, projectID string) (*schema.File, error)
	Update(file *schema.File) error
	UpdateFields(fileID string, fields map[string]interface{}) error
	Delete(fileID, directoryID, projectID string) (*schema.File, error)
}

// Groups allows manipulation and access to groups.
type Groups interface {
	ByID(id string) (*schema.Group, error)
	ForOwner(owner string) ([]schema.Group, error)
}

// Uploads allows manipulation and access to upload requests.
type Uploads interface {
	ByID(id string) (*schema.Upload, error)
	Insert(upload *schema.Upload) (*schema.Upload, error)
	Update(upload *schema.Upload) error
	ForUser(user string) ([]schema.Upload, error)
	Delete(uploadID string) error
}

// Projects is an interface describing access to projects in the system.
type Projects interface {
	ByID(id string) (*schema.Project, error)
	HasDirectory(projectID, directoryID string) bool
}

// Dirs is an interface describing access to directories in the system.
type Dirs interface {
	ByID(id string) (*schema.Directory, error)
}
