package dai

import (
	"testing"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/testutil"
	"github.com/stretchr/testify/require"
)

var ruploads = NewRUploads(testutil.RSession())

func TestRUploadsForUser(t *testing.T) {
	// Test no user
	uploads, err := ruploads.ForUser("no-such-user")
	require.NotNil(t, err)
	require.Nil(t, uploads)
	require.Equal(t, err, app.ErrNotFound, "Error unexpected value: %s", err)

	// Insert upload for user and find for that user
	upload := schema.CUpload().
		Owner("test@mc.org").
		Project("test", "test").
		Directory("test", "test").
		Create()
	newUpload, err := ruploads.Insert(&upload)
	require.Nil(t, err)
	require.NotNil(t, newUpload)
	require.Equal(t, newUpload.Owner, "test@mc.org")

	uploads, err = ruploads.ForUser("test@mc.org")
	require.Nil(t, err)
	require.NotNil(t, uploads)
	require.Equal(t, len(uploads), 1)

	err = ruploads.Delete(uploads[0].ID)
	require.Nil(t, err)
}

func TestRUploadsUpdate(t *testing.T) {
	upload := schema.CUpload().
		Owner("test@mc.org").
		Project("test", "test").
		Create()
	newUpload, err := ruploads.Insert(&upload)
	require.Nil(t, err)
	require.NotNil(t, newUpload)
	require.Equal(t, newUpload.Owner, "test@mc.org")

	newUpload.ProjectName = "changedName"
	err = ruploads.Update(newUpload)
	require.Nil(t, err)

	newUpload, err = ruploads.ByID(newUpload.ID)
	require.Nil(t, err)
	require.NotNil(t, newUpload)
	require.Equal(t, newUpload.ProjectName, "changedName")

	err = ruploads.Delete(newUpload.ID)
	require.Nil(t, err)
}
