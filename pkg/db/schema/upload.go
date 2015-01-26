package schema

import "time"

// FileUpload is the tracking information for an individual file upload.
type FileUpload struct {
	Name        string    `gorethink:"name"`         // File name on remote system
	Checksum    string    `gorethink:"checksum"`     // Computed file checksum
	Size        int64     `gorethink:"size"`         // Size of file on remote system
	Birthtime   time.Time `gorethink:"birthtime"`    // When was FileUpload started
	MTime       time.Time `gorethink:"mtime"`        // Last time this entry was modified
	RemoteMTime time.Time `gorethink:"remote_mtime"` // CTime of the remote file
	ChunkSize   int       `gorethink:"chunk_size"`   // Chunk transfer size
	ChunkCount  int       `gorethink:"chunk_count"`  // Number of chunks expected
	ChunkHashes []string  `gorethink:"chunk_hashes"` // Hash for each uploaded chunk
}

// A Upload models a user upload request. It allows for users to restart
// upload requests.
type Upload struct {
	ID            string     `gorethink:"id,omitempty"`
	Owner         string     `gorethink:"owner"`          // Who started the upload
	DirectoryID   string     `gorethink:"directory_id"`   // Directory to upload to
	DirectoryName string     `gorethink:"directory_name"` // Name of directory
	ProjectID     string     `gorethink:"project_id"`     // Project to upload to
	ProjectName   string     `gorethink:"project_name"`   // Name of project
	Birthtime     time.Time  `gorethink:"birthtime"`      // When was upload started
	Host          string     `gorethink:"host"`           // Host requesting the upload
	File          FileUpload `gorethink:"file"`           // File being uploaded
}

// uploadCreater allows for chaining creation of an upload.
type uploadCreater struct {
	upload Upload
}

// CUpload creates a new uploadCreater.
func CUpload() *uploadCreater {
	u := &uploadCreater{}
	now := time.Now()
	u.upload.Birthtime = now
	u.upload.File.Birthtime = now
	u.upload.File.MTime = now
	u.upload.File.RemoteMTime = now
	return u
}

// Owner sets the Upload owner field.
func (c *uploadCreater) Owner(owner string) *uploadCreater {
	c.upload.Owner = owner
	return c
}

// Directory sets the Upload DirectoryID and DirectoryName fields.
func (c *uploadCreater) Directory(id, name string) *uploadCreater {
	c.upload.DirectoryID = id
	c.upload.DirectoryName = name
	return c
}

// Project sets the Upload ProjectID and ProjectName fields.
func (c *uploadCreater) Project(id, name string) *uploadCreater {
	c.upload.ProjectID = id
	c.upload.ProjectName = name
	return c
}

// Birthtime sets the Upload Birthtime field.
func (c *uploadCreater) Birthtime(birthtime time.Time) *uploadCreater {
	c.upload.Birthtime = birthtime
	return c
}

// Host sets the Upload Host field.
func (c *uploadCreater) Host(host string) *uploadCreater {
	c.upload.Host = host
	return c
}

// FName sets the Upload.File.Name field.
func (c *uploadCreater) FName(name string) *uploadCreater {
	c.upload.File.Name = name
	return c
}

// FChecksum sets the Upload.File.Checksum field.
func (c *uploadCreater) FChecksum(checksum string) *uploadCreater {
	c.upload.File.Checksum = checksum
	return c
}

// FSize sets the Upload.File.Size field.
func (c *uploadCreater) FSize(size int64) *uploadCreater {
	c.upload.File.Size = size
	return c
}

// FTime sets the Upload.File.Birthtime and Upload.File.MTime fields.
func (c *uploadCreater) FTime(t time.Time) *uploadCreater {
	c.upload.File.Birthtime = t
	c.upload.File.MTime = t
	return c
}

// FRemoteCTime sets the Upload.File.RemoteCTime field.
func (c *uploadCreater) FRemoteMTime(t time.Time) *uploadCreater {
	c.upload.File.RemoteMTime = t
	return c
}

// FChunk sets the Upload.File.ChunkSize and the Upload.File.ChunkCount fields.
func (c *uploadCreater) FChunk(size, count int) *uploadCreater {
	c.upload.File.ChunkSize = size
	c.upload.File.ChunkCount = count
	return c
}

// Create creates a new Upload instance from the values given to the
// the creater.
func (c *uploadCreater) Create() Upload {
	return c.upload
}
