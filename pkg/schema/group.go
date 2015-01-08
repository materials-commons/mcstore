package schema

import (
	"time"
)

// Group models users groups and access permissions to user data.
type Group struct {
	ID          string    `gorethink:"id,omitempty"`
	Owner       string    `gorethink:"owner"`
	Name        string    `gorethink:"name"`
	Description string    `gorethink:"description"`
	Birthtime   time.Time `gorethink:"birthtime"`
	MTime       time.Time `gorethink:"mtime"`
	Access      string    `gorethink:"access"`
	Users       []string  `gorethink:"users"`
}

// NewGroup creates a new UserGroup instance.
func NewGroup(owner, name string) Group {
	now := time.Now()
	return Group{
		Owner:       owner,
		Name:        name,
		Description: name,
		Access:      "private",
		Birthtime:   now,
		MTime:       now,
	}
}
