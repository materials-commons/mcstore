package schema

import "time"

// FileUpload is the tracking information for an individual file upload.
type FileUpload struct {
	Name        string    `gorethink:"name"`         // File name on remote system
	Checksum    string    `gorethink:"checksum"`     // Computed file checksum
	Size        int64     `gorethink:"size"`         // Size of file on remote system
	Birthtime   time.Time `gorethink:"birthtime"`    // When was FileUpload started
	MTime       time.Time `gorethink:"mtime"`        // Last time this entry was modified
	RemoteCTime time.Time `gorethink:"remote_ctime"` // CTime of the remote file
	ChunkSize   int       `gorethink:"chunk_size"`   // Chunk transfer size
	ChunkCount  int       `gorethink:"chunk_count"`  // Number of chunks expected
	ChunkHashes []string  `gorethink:"chunk_hashes"` // Hash for each uploaded chunk
}

// A Upload models a user upload request. It allows for users to restart
// upload requests.
type Upload struct {
	ID          string     `gorethink:"id,omitempty"`
	Owner       string     `gorethink:"owner"`        // Who started the upload
	DirectoryID string     `gorethink:"directory_id"` // Directory to upload to
	ProjectID   string     `gorethink:"project_id"`   // Project to upload to
	Birthtime   time.Time  `gorethink:"birthtime"`    // When was upload started
	Host        string     `gorethink:"host"`         // Host requesting the upload
	File        FileUpload `gorethink:"file"`         // File being uploaded
}
