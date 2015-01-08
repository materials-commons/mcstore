package schema

import (
	"time"
)

// Project models a users project. A project is an instance of a users workspace
// where they conduct their research. A project can be shared.
type Project struct {
	ID          string    `gorethink:"id,omitempty"`
	Name        string    `gorethink:"name"`
	Description string    `gorethink:"description"`
	DataDir     string    `gorethink:"datadir" db:"-"`
	Owner       string    `gorethink:"owner" db:"-"`
	Birthtime   time.Time `gorethink:"birthtime"`
	MTime       time.Time `gorethink:"mtime"`
	Notes       []Note    `gorethink:"notes" db:"-"`
	Tags        []string  `gorethink:"tags" db:"-"`
	Reviews     []string  `gorethink:"reviews" db:"-"`
	MyTags      []string  `gorethink:"mytags" db:"-"`
}

// NewProject creates a new Project instance.
func NewProject(name, datadir, owner string) Project {
	now := time.Now()
	return Project{
		Name:      name,
		DataDir:   datadir,
		Owner:     owner,
		Birthtime: now,
		MTime:     now,
	}
}
