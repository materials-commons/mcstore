package dai

import (
	"testing"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/test"
	"github.com/stretchr/testify/require"
)

var rfiles = NewRFiles(test.RSession())

func TestRFilesByID(t *testing.T) {
	// Test existing
	f, err := rfiles.ByID("testfile.txt")
	require.Nil(t, err, "Unable retrieve existing file: %s", err)
	require.NotNil(t, f, "Found file, but returned nil for entry")
	require.Equal(t, f.ID, "testfile.txt", "Retrieved wrong file %#v", f)

	// Test non-existant
	f, err = rfiles.ByID("does-not-exist")
	require.Equal(t, err, app.ErrNotFound, "Found file that doesn't exist")
	require.Nil(t, f, "Returned file entry rather than nil %#v", f)
}
