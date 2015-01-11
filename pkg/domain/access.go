package domain

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// TODO: Group caching
// TODO: cache reloading

// Access validates access to data. It checks if a user
// has been given permission to access a particular item.
type Access struct {
	groups dai.Groups
	files  dai.Files
	users  dai.Users
}

// NewAccess creates a new Access.
func NewAccess(groups dai.Groups, files dai.Files, users dai.Users) *Access {
	return &Access{
		groups: groups,
		files:  files,
		users:  users,
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
func (a *Access) AllowedByOwner(owner, user string) bool {
	// Check if user and file owner are the same, or the user is
	// in the admin group.
	if user == owner || a.isAdmin(user) {
		return true
	}

	// Get the owners groups
	groups, err := a.groups.ForOwner(owner)
	if err != nil {
		// Some sort of error occurred, assume no access
		return false
	}

	// For each group go through its list of users and see if
	// they match the requesting user. If there is a match
	// then the owner has given access to the user.
	for _, group := range groups {
		users := group.Users
		for _, u := range users {
			if u == user {
				return true
			}
		}
	}

	return false
}

// isAdmin checks if a user is in the admin group.
func (a *Access) isAdmin(user string) bool {
	group, err := a.groups.ByID("admin")
	if err != nil {
		return false
	}

	for _, admin := range group.Users {
		if admin == user {
			return true
		}
	}

	return false
}

// GetFile will validate access to a file. Rather than taking a user,
// it takes an apikey and looks up the user. It returns the file if
// access has been granted, otherwise it returns the erro ErrNoAccess.
func (a *Access) GetFile(apikey, fileID string) (*schema.File, error) {
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

	if !a.AllowedByOwner(file.Owner, user.ID) {
		app.Log.Info("Access denied", "fileid", file.ID, "user", user.ID)
		return nil, app.ErrNoAccess
	}

	return file, nil
}
