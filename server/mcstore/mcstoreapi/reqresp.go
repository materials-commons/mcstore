package mcstoreapi

import "time"

// UploadEntry is a client side representation of an upload.
type UploadEntry struct {
	RequestID   string    `json:"request_id"`
	FileName    string    `json:"filename"`
	DirectoryID string    `json:"directory_id"`
	ProjectID   string    `json:"project_id"`
	Size        int64     `json:"size"`
	Host        string    `json:"host"`
	Checksum    string    `json:"checksum"`
	Birthtime   time.Time `json:"birthtime"`
}

// CreateRequest describes the JSON request a client will send
// to create a new upload request.
type CreateUploadRequest struct {
	ProjectID   string `json:"project_id"`
	DirectoryID string `json:"directory_id"`
	FileName    string `json:"filename"`
	FileSize    int64  `json:"filesize"`
	ChunkSize   int32  `json:"chunk_size"`
	FileMTime   string `json:"filemtime"`
	Checksum    string `json:"checksum"`
}

// uploadCreateResponse is the format of JSON sent back containing
// the upload request ID.
type CreateUploadResponse struct {
	RequestID     string `json:"request_id"`
	StartingBlock uint   `json:"starting_block"`
}

type UploadChunkResponse struct {
	FileID string `json:"file_id"`
	Done   bool   `json:"done"`
}

// CreateProjectRequest requests that a project be created. If MustNotExist
// is true, then the given project must not already exist. Existence is
// determined by the project name for that user.
type CreateProjectRequest struct {
	Name         string `json:"name"`
	MustNotExist bool   `json:"must_not_exist"`
}

// CreateProjectResponse returns the created project. If the project was an
// existing project and no new project was created then the Existing flag
// will be set to false.
type CreateProjectResponse struct {
	ProjectID string `json:"project_id"`
	Existing  bool   `json:"existing"`
}

// GetDirectoryRequest is a request to get a directory for a project. The
// directory lookup is by path within the context of the given project.
type GetDirectoryRequest struct {
	Path      string
	ProjectID string
}

// GetDirectoryResponse returns the directory id for a directory
// path for a given project.
type GetDirectoryResponse struct {
	DirectoryID string `json:"directory_id"`
	Path        string `json:"path"`
}
