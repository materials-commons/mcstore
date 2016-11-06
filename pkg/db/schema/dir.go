package schema

import (
	"time"
)

// Directory models a directory of user files. A dir is an abstract representation
// of a users file system directory plus the metadata needed by the system.
type Directory struct {
	ID        string    `gorethink:"id,omitempty" json:"id"`
	Type      string    `gorethink:"otype" json:"otype"`
	Owner     string    `gorethink:"owner" json:"owner"`
	Name      string    `gorethink:"name" json:"name"`
	Project   string    `gorethink:"project" json:"project"`
	Parent    string    `gorethink:"parent" json:"parent"`
	Birthtime time.Time `gorethink:"birthtime" json:"birthtime"`
	MTime     time.Time `gorethink:"mtime" json:"mtime"`
	ATime     time.Time `gorethink:"atime" json:"mtime"`
}

// NewDirectory creates a new Directory instance.
func NewDirectory(name, owner, project, parent string) Directory {
	now := time.Now()
	return Directory{
		Type:      "datadir",
		Owner:     owner,
		Name:      name,
		Project:   project,
		Parent:    parent,
		Birthtime: now,
		MTime:     now,
		ATime:     now,
	}
}
