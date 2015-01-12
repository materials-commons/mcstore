package domain

import (
	"testing"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai/mocks"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/stretchr/testify/require"
)

func TestIsAdmin(t *testing.T) {
	musers := mocks.NewMUsers()
	mfiles := mocks.NewMFiles()
	mgroups := mocks.NewMGroups()
	a := NewAccess(mgroups, mfiles, musers)

	var group schema.Group
	group.Users = append(group.Users, "test@mc.org")

	// Test in admin group
	mgroups.On("ByID", "admin").Return(&group, nil)
	require.True(t, a.isAdmin("test@mc.org"), "Expected test@mc.org to be admin")

	// Test not in admin group
	mgroups.On("ByID", "admin").Return(&group, app.ErrNotFound)
	require.False(t, a.isAdmin("test@mc.org"), "Unexpected test@mc.org is admin")
}

func TestAllowedByOwner(t *testing.T) {
	musers := mocks.NewMUsers()
	mfiles := mocks.NewMFiles()
	mgroups := mocks.NewMGroups()
	a := NewAccess(mgroups, mfiles, musers)

	var group schema.Group
	group.Users = append(group.Users, "test1@mc.org")

	var groupsEmpty []schema.Group
	var groupsWithUser []schema.Group
	groupsWithUser = append(groupsWithUser, group)

	// Test no admin group and not allowed by user
	mgroups.On("ByID", "admin").Return(&group, app.ErrNotFound)
	mgroups.On("ForOwner", "test@mc.org").Return(groupsEmpty, nil)
	require.False(t, a.AllowedByOwner("test@mc.org", "test1@mc.org"), "Should not have allowed access")

	// Test in admin group
	mgroups.On("ByID", "admin").Return(&group, nil)
	require.True(t, a.AllowedByOwner("test@mc.org", "test1@mc.org"), "Should have allowed access")

	// Test owner did not give access
	mgroups.On("ByID", "admin").Return(&group, app.ErrNotFound)
	require.False(t, a.AllowedByOwner("test@mc.org", "test1@mc.org"), "Should not have allowed access")

	// Test owner did give access
	mgroups.On("ForOwner", "test@mc.org").Return(groupsWithUser, nil)
	require.True(t, a.AllowedByOwner("test@mc.org", "test1@mc.org"), "Should have allowed access")
}

func TestGetFile(t *testing.T) {
	musers := mocks.NewMUsers()
	mfiles := mocks.NewMFiles()
	mgroups := mocks.NewMGroups()
	a := NewAccess(mgroups, mfiles, musers)

	var group schema.Group
	group.Users = append(group.Users, "test1@mc.org")

	var groupsEmpty []schema.Group
	var groupsWithUser []schema.Group
	groupsWithUser = append(groupsWithUser, group)

	_ = groupsEmpty

	// Test bad apikey
	var nilUser *schema.User = nil
	musers.On("ByAPIKey", "abc123").Return(nilUser, app.ErrNotFound)
	f, err := a.GetFile("abc123", "fileid")
	require.Equal(t, err, app.ErrNoAccess, "Incorrect error %s", err)
	require.Nil(t, f, "File should have been nil")

	// Test no such file
	var nilFile *schema.File = nil
	var user = schema.NewUser("test1", "test1@mc.org", "password", "abc123")
	mfiles.On("ByID", "fileid").Return(nilFile, app.ErrNotFound)
	musers.On("ByAPIKey", "abc123").Return(&user, nil)
	f, err = a.GetFile("abc123", "fileid")
	require.Equal(t, err, app.ErrNoAccess, "Incorrect error %s", err)
	require.Nil(t, f, "File should have been nil")

	// Test admin access
	var file = schema.NewFile("fileid.txt", "test@mc.org")
	mfiles.On("ByID", "fileid").Return(&file, nil)
	mgroups.On("ByID", "admin").Return(&group, nil)
	f, err = a.GetFile("abc123", "fileid")
	require.Nil(t, err, "Wrong error %s", err)
	require.NotNil(t, f, "Got nil file")

	// Test access granted
	// First remove admin access so it doesn't override
	mgroups.On("ByID", "admin").Return(&group, app.ErrNotFound)
	mgroups.On("ForOwner", "test@mc.org").Return(groupsWithUser, nil)
	f, err = a.GetFile("abc123", "fileid")
	require.Nil(t, err, "Wrong error %s", err)
	require.NotNil(t, f, "Got nil file")

	// Test access not granted
	// First add a user not in the groups
	user = schema.NewUser("test2", "test2@mc.org", "password", "def456")
	musers.On("ByAPIKey", "def456").Return(&user, nil)
	f, err = a.GetFile("abc123", "fileid")
	require.Equal(t, err, app.ErrNoAccess, "Wrong error %s", err)
	require.Nil(t, f, "Got file, should have been nil")
}
