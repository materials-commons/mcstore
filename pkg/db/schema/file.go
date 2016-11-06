package schema

import (
	"time"
)

type fileFields int

// File contains methods to get the json fields names for a File
var FileFields fileFields

func (f fileFields) ID() string          { return "id" }
func (f fileFields) Current() string     { return "current" }
func (f fileFields) Name() string        { return "name" }
func (f fileFields) Birthtime() string   { return "birthtime" }
func (f fileFields) MTime() string       { return "mtime" }
func (f fileFields) ATime() string       { return "atime" }
func (f fileFields) Description() string { return "description" }
func (f fileFields) MediaType() string   { return "mediatype" }
func (f fileFields) Owner() string       { return "owner" }
func (f fileFields) Checksum() string    { return "checksum" }
func (f fileFields) Size() string        { return "size" }
func (f fileFields) Uploaded() string    { return "uploaded" }
func (f fileFields) Parent() string      { return "parent" }
func (f fileFields) UsesID() string      { return "usesid" }

// MediaType describes the mime media type and its description.
type MediaType struct {
	Mime        string `gorethink:"mime" json:"mime"`               // The MIME type
	Description string `gorethink:"description" json:"description"` // Description of MIME type

	// MIME Description translated to human readable format
	MimeDescription string `gorethink:"mime_description" json:"mime_description"`
}

// File models a user file. A datafile is an abstract representation of a real file
// plus the attributes that we need in our model for access, and other metadata.
type File struct {
	ID          string    `gorethink:"id,omitempty" json:"id"`         // Primary key.
	Type        string    `gorethink:"otype" json:"otype"`             // Type
	Current     bool      `gorethink:"current" json:"current"`         // Is this the most current version.
	Name        string    `gorethink:"name" json:"name"`               // Name of file.
	Path        string    `gorethink:"path,omitempty" json:"path"`     // Directory path where file resides.
	Birthtime   time.Time `gorethink:"birthtime" json:"birthtime"`     // Creation time.
	MTime       time.Time `gorethink:"mtime" json:"mtime"`             // Modification time.
	ATime       time.Time `gorethink:"atime" json:"atime"`             // Last access time.
	Description string    `gorethink:"description" json:"description"` // Description of file
	MediaType   MediaType `gorethink:"mediatype" json:"mediatype"`     // File media type and description
	Owner       string    `gorethink:"owner" json:"owner"`             // Who owns the file.
	Checksum    string    `gorethink:"checksum" json:"checksum"`       // MD5 Hash.
	Size        int64     `gorethink:"size" json:"size"`               // Size of file.
	Uploaded    int64     `gorethink:"uploaded" json:"-"`              // Number of bytes uploaded. When Size != Uploaded file is only partially uploaded.
	Parent      string    `gorethink:"parent" json:"parent"`           // If there are multiple ids then parent is the id of the previous version.
	UsesID      string    `gorethink:"usesid" json:"usesid"`           // If file is a duplicate, then usesid points to the real file. This allows multiple files to share a single physical file.
}

// NewFile creates a new File instance.
func NewFile(name, owner string) File {
	now := time.Now()
	return File{
		Name:        name,
		Owner:       owner,
		Description: "",
		Birthtime:   now,
		MTime:       now,
		ATime:       now,
		Current:     true,
		Type:        "datafile",
	}
}

// FileID returns the id to use for the file. Because files can be duplicates, all
// duplicates are stored under a single ID. UsesID is set to the ID that an entry
// points to when it is a duplicate.
func (f *File) FileID() string {
	if f.UsesID != "" {
		return f.UsesID
	}

	return f.ID
}

// private type to hang helper methods off of
type fs struct{}

// Files gives access to helper routines that work on lists of files.
var Files fs

// Find will return a matching File in a list of files when the match func returns true.
func (f fs) Find(files []File, match func(f File) bool) *File {
	for _, file := range files {
		if match(file) {
			return &file
		}
	}

	return nil
}
