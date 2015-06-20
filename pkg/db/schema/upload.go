package schema

import (
	"time"

	"github.com/willf/bitset"
)

// FileUpload is the tracking information for an individual file upload.
type FileUpload struct {
	Name        string         `gorethink:"name"`         // File name on remote system
	Checksum    string         `gorethink:"checksum"`     // Computed file checksum
	Size        int64          `gorethink:"size"`         // Size of file on remote system
	Birthtime   time.Time      `gorethink:"birthtime"`    // When was FileUpload started
	MTime       time.Time      `gorethink:"mtime"`        // Last time this entry was modified
	RemoteMTime time.Time      `gorethink:"remote_mtime"` // CTime of the remote file
	ChunkSize   int            `gorethink:"chunk_size"`   // Chunk transfer size
	ChunkCount  int            `gorethink:"chunk_count"`  // Number of chunks expected
	Blocks      *bitset.BitSet // Block state. If set block as has been uploaded
	BitString   []byte
}

// A Upload models a user upload request. It allows for users to restart
// upload requests.
type Upload struct {
	ID            string     `gorethink:"id,omitempty"`
	Owner         string     `gorethink:"owner"`          // Who started the upload
	DirectoryID   string     `gorethink:"directory_id"`   // Directory to upload to
	DirectoryName string     `gorethink:"directory_name"` // Name of directory
	ProjectID     string     `gorethink:"project_id"`     // Project to upload to
	ProjectOwner  string     `gorethink:"project_owner"`  // Owner of project
	ProjectName   string     `gorethink:"project_name"`   // Name of project
	Birthtime     time.Time  `gorethink:"birthtime"`      // When was upload started
	Host          string     `gorethink:"host"`           // Host requesting the upload
	File          FileUpload `gorethink:"file"`           // File being uploaded
}

// SetFBlocks sets the blocks and BitString. It does nothing if blocks is nil.
func (u *Upload) SetFBlocks(blocks *bitset.BitSet) {
	if blocks != nil {
		u.File.Blocks = blocks
		u.File.BitString, _ = blocks.MarshalJSON()
	}
}

// private type to hang helper methods off of.
type uploadList struct{}

// Uploads gives access to helper routines that work on lists of uploads.
var Uploads uploadList

// Find will return a matching Upload in a list of uploads when the match func returns true.
func (u uploadList) Find(uploads []Upload, match func(upload Upload) bool) *Upload {
	for _, upload := range uploads {
		if match(upload) {
			return &upload
		}
	}

	return nil
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
	u.upload.File.Blocks = bitset.New(1)
	u.upload.File.BitString, _ = u.upload.File.Blocks.MarshalJSON()
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

// ProjectOnwer sets the Upload ProjectOwner field.
func (c *uploadCreater) ProjectOwner(owner string) *uploadCreater {
	c.upload.ProjectOwner = owner
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

// FBlocks sets the Upload.File.Blocks field. It also sets the BitString
// field with a byte array representation of the bitset.
func (c *uploadCreater) FBlocks(blocks *bitset.BitSet) *uploadCreater {
	c.upload.File.Blocks = blocks
	c.upload.File.BitString, _ = blocks.MarshalJSON()
	return c
}

// Create creates a new Upload instance from the values given to the
// the creater.
func (c *uploadCreater) Create() Upload {
	return c.upload
}
