package schema

import (
	"time"
)

type MediaTypeSummary struct {
	Count       int    `gorethink:"count" json:"count"`
	Description string `gorethink:"description" json:"description"`
	Size        int64  `gorethink:"size" json:"size"`
}

// Project models a users project. A project is an instance of a users workspace
// where they conduct their research. A project can be shared.
type Project struct {
	ID          string                      `gorethink:"id,omitempty"`
	Type        string                      `gorethink:"_type" json:"_type"`
	Name        string                      `gorethink:"name"`
	Description string                      `gorethink:"description"`
	DataDir     string                      `gorethink:"datadir" db:"-"`
	Owner       string                      `gorethink:"owner" db:"-"`
	Birthtime   time.Time                   `gorethink:"birthtime"`
	MTime       time.Time                   `gorethink:"mtime"`
	MediaTypes  map[string]MediaTypeSummary `gorethink:"mediatypes" json:"mediatypes"`
	Size        int64                       `gorethink:"size" json:"size"`
}

// NewProject creates a new Project instance.
func NewProject(name, owner string) Project {
	now := time.Now()
	return Project{
		Name:      name,
		Type:      "project",
		Owner:     owner,
		Birthtime: now,
		MTime:     now,
	}
}
