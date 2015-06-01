package dai

import (
	"testing"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/testutil"
	"github.com/stretchr/testify/require"
)

var rusers = NewRUsers(testutil.RSession())

func TestRUsersByID(t *testing.T) {
	// Test existing
	u, err := rusers.ByID("test@mc.org")
	require.NotNil(t, u, "Unable to retrieve existing user test@mc.org %s", err)
	require.Equal(t, u.ID, "test@mc.org", "Wrong user retrieved expected test@mc.org, got user %#v", u)

	// Test non-existant user
	u, err = rusers.ByID("does@not.exist")
	require.Equal(t, err, app.ErrNotFound)
	require.Nil(t, u, "Retrieved non existant user does@not.exist")
}

func TestRUsersByAPIKey(t *testing.T) {
	// Test existing
	u, err := rusers.ByAPIKey("test")
	require.Nil(t, err, "Failed to retrieve apikey test %s", err)
	require.Equal(t, u.ID, "test@mc.org", "Wrong user with apikey test: %#v", u)

	// Test non-existant key
	u, err = rusers.ByAPIKey("no-such-key")
	require.Equal(t, err, app.ErrNotFound, "Retrieved key that does not exist, got %#v", u)
	require.Nil(t, u, "Retrieved user for bad key does@not.exist")
}
