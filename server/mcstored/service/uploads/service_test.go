package uploads

import (
	"testing"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/materials-commons/mcstore/test"
	"github.com/stretchr/testify/require"
)

var (
	users    = dai.NewRUsers(test.RSession())
	files    = dai.NewRFiles(test.RSession())
	groups   = dai.NewRGroups(test.RSession())
	dirs     = dai.NewRDirs(test.RSession())
	projects = dai.NewRProjects(test.RSession())
	uploads  = dai.NewRUploads(test.RSession())
)

func TestCreateHasAccess(t *testing.T) {
	access := domain.NewAccess(groups, files, users)
	s := NewCreateServiceFrom(dirs, projects, uploads, access)

	// Test with admin
	cf := CreateRequest{
		ProjectID:   "test",
		DirectoryID: "test",
		User:        "test@mc.org",
		Host:        "host",
	}
	upload, err := s.Create(cf)
	require.Nil(t, err)
	require.NotNil(t, upload)
	err = uploads.Delete(upload.ID)
	require.Nil(t, err)

	// Test with group access
	cf.User = "test1@mc.org"
	upload, err = s.Create(cf)
	require.Nil(t, err)
	require.NotNil(t, upload)
	err = uploads.Delete(upload.ID)
	require.Nil(t, err)
}

func TestCreateNoAccess(t *testing.T) {
	access := domain.NewAccess(groups, files, users)
	s := NewCreateServiceFrom(dirs, projects, uploads, access)

	// Test with admin
	cf := CreateRequest{
		ProjectID:   "test",
		DirectoryID: "test",
		User:        "test2@mc.org",
		Host:        "host",
	}
	upload, err := s.Create(cf)
	require.NotNil(t, err)
	require.Equal(t, err, app.ErrNoAccess)
	require.Nil(t, upload)
}

func TestCreateInvalidRequest(t *testing.T) {
	access := domain.NewAccess(groups, files, users)
	s := NewCreateServiceFrom(dirs, projects, uploads, access)

	cf := CreateRequest{
		ProjectID:   "test",
		DirectoryID: "not-exist",
		User:        "test@mc.org",
		Host:        "host",
	}

	upload, err := s.Create(cf)
	require.NotNil(t, err)
	require.Nil(t, upload)
}
