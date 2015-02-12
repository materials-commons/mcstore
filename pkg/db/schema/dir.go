package schema

import "time"

// Directory models a directory of user files. A dir is an abstract representation
// of a users file system directory plus the metadata needed by the system.
type Directory struct {
	ID        string    `gorethink:"id,omitempty"`
	Type      string    `gorethink:"_type"`
	Owner     string    `gorethink:"owner"`
	Name      string    `gorethink:"name"`
	Project   string    `gorethink:"project"`
	Parent    string    `gorethink:"parent"`
	Birthtime time.Time `gorethink:"birthtime"`
	MTime     time.Time `gorethink:"mtime"`
	ATime     time.Time `gorethink:"atime"`
	Files     []File    `gorethink:"datafiles,omitempty"`
}

// NewDirectory creates a new Directory instance.
func NewDirectory(name, owner, project, parent string) Directory {
	now := time.Now()
	return Directory{
		Owner:     owner,
		Name:      name,
		Project:   project,
		Parent:    parent,
		Birthtime: now,
		MTime:     now,
		ATime:     now,
	}
}
