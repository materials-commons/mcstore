package domain

import "github.com/materials-commons/mcstore/pkg/db/dai"

// TODO: Group caching
// TODO: cache reloading

type Access struct {
	groups dai.Groups
}

func NewAccess(groups dai.Groups) *Access {
	return &Access{
		groups: groups,
	}
}

// AllowedByOwner checks to see if the user making the request has access to the
// particular item. Access is determined as follows:
// 1. If the user and the owner of the item are the same return true (has access).
// 2. Get a list of all the users groups for the item's owner.
//    For each user in the user group see if the requesting user
//    is included. If so then return true (has access).
// 3. None of the above matched - return false (no access)
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
