package domain

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// TODO: Add Redis as a store for this

type Access interface {
	AllowedByOwner(projectID, user string) bool
	GetFile(apikey, fileID string) (*schema.File, error)
}

// access validates access to data. It checks if a user
// has been given permission to access a particular item.
type access struct {
	projects dai.Projects
	files    dai.Files
	users    dai.Users
}

// NewAccess creates a new Access.
func NewAccess(projects dai.Projects, files dai.Files, users dai.Users) *access {
	return &access{
		projects: projects,
		files:    files,
		users:    users,
	}
}

// AllowedByOwner checks to see if the user making the request has access to the
// particular item. Access is determined as follows:
// 1. If the user and the owner of the item are the same
//    or the user is in the admin group return true (has access).
// 2. Get a list of all the users groups for the item's owner.
//    For each user in the user group see if the requesting user
//    is included. If so then return true (has access).
// 3. None of the above matched - return false (no access).
func (a *access) AllowedByOwner(projectID, user string) bool {
	u, err := a.users.ByID(user)
	if err != nil {
		return false
	}

	if u.Admin {
		return true
	}

	accessList, err := a.projects.AccessList(projectID)
	if err != nil {
		return false
	}
	for _, entry := range accessList {
		if user == entry.UserID {
			return true
		}
	}
	return false
}

// GetFile will validate access to a file. Rather than taking a user,
// it takes an apikey and looks up the user. It returns the file if
// access has been granted, otherwise it returns the erro ErrNoAccess.
func (a *access) GetFile(apikey, fileID string) (*schema.File, error) {
	user, err := a.users.ByAPIKey(apikey)
	if err != nil {
		// log error here
		app.Log.Error("User lookup failed", "error", err, "apikey", apikey)
		return nil, app.ErrNoAccess
	}

	file, err := a.files.ByID(fileID)
	if err != nil {
		app.Log.Error("File lookup failed", "error", err, "fileid", fileID)
		return nil, app.ErrNoAccess
	}

	project, err := a.files.GetProject(fileID)
	if err != nil {
		app.Log.Error("Project lookup for file failed", "error", err, "fileid", fileID)
		return nil, app.ErrNoAccess
	}

	if !a.AllowedByOwner(project.ID, user.ID) {
		app.Log.Info("Access denied", "fileid", file.ID, "user", user.ID, "projectid", project.ID)
		return nil, app.ErrNoAccess
	}

	return file, nil
}
