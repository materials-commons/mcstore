package dai

import (
	"testing"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/test"
	"github.com/stretchr/testify/require"
)

var rgroups = NewRGroups(test.RSession())

func TestRGroupsByID(t *testing.T) {
	// Test existing group
	g, err := rgroups.ByID("test")
	require.Nil(t, err, "Unable to find group test, err %s", err)
	require.NotNil(t, g, "Found group, but returned nil")
	require.Equal(t, g.Name, "test", "Retrieved wrong group %#v", g)

	// Test doesn't exist
	g, err = rgroups.ByID("does-not-exist")
	require.Equal(t, err, app.ErrNotFound, "Unexpected err value: %s, expected ErrNotFound", err)
	require.Nil(t, g, "Group returned, even though not found: %#v", g)
}

func TestRGroupsForOwner(t *testing.T) {
	// Test existing owner
	groups, err := rgroups.ForOwner("test@mc.org")
	require.Nil(t, err, "Unable to find groups for owner 'test@mc.org': %s", err)
	require.NotNil(t, groups, "Found groups for owner, but returned nil")

	// Test doesn't exist
	groups, err = rgroups.ForOwner("does-not-exist")
	require.Equal(t, err, app.ErrNotFound, "Err not equal to ErrNotFound: %s", err)
	require.Nil(t, groups, "No group, but got non nil value: %#v", groups)
}
