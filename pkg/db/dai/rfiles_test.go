package dai

import (
	"testing"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/testutil"
	"github.com/stretchr/testify/require"
)

var rfiles = NewRFiles(testutil.RSession())

func TestRFilesByID(t *testing.T) {
	// Test existing
	f, err := rfiles.ByID("testfile.txt")
	require.Nil(t, err, "Unable retrieve existing file: %s", err)
	require.NotNil(t, f, "Found file, but returned nil for entry")
	require.Equal(t, f.ID, "testfile.txt", "Retrieved wrong file %#v", f)

	// Test no such file id
	f, err = rfiles.ByID("does-not-exist")
	require.Equal(t, err, app.ErrNotFound, "Found file that doesn't exist")
	require.Nil(t, f, "Returned file entry rather than nil %#v", f)
}

func TestRFilesInsertIDSet(t *testing.T) {
	file := schema.NewFile("test1.txt", "test@mc.org")
	file.ID = "test1.txt" // Explicitly set the ID

	newFile, err := rfiles.Insert(&file, "test", "test")
	require.Nil(t, err)
	require.NotNil(t, newFile)
	require.Equal(t, newFile.ID, "test1.txt")

	// Check that the join tables were updated.
	session := testutil.RSession()
	var p2df []schema.Project2DataFile
	rql := r.Table("project2datafile").Filter(r.Row.Field("datafile_id").Eq(file.ID))
	err = model.ProjectFiles.Qs(session).Rows(rql, &p2df)
	require.Nil(t, err)
	require.Equal(t, len(p2df), 1)

	var dir2df []schema.DataDir2DataFile
	rql = r.Table("datadir2datafile").Filter(r.Row.Field("datafile_id").Eq(file.ID))
	err = model.DirFiles.Qs(session).Rows(rql, &dir2df)
	require.Nil(t, err)
	require.Equal(t, len(dir2df), 1)

	deleteFile(file.ID)
}

func TestRFilesInsertNoIDSet(t *testing.T) {
	file := schema.NewFile("test1.txt", "test@mc.org")

	newFile, err := rfiles.Insert(&file, "test", "test")
	require.Nil(t, err)
	require.NotNil(t, newFile)

	// Check that the join tables were updated.
	session := testutil.RSession()
	var p2df []schema.Project2DataFile
	rql := r.Table("project2datafile").Filter(r.Row.Field("datafile_id").Eq(newFile.ID))
	err = model.ProjectFiles.Qs(session).Rows(rql, &p2df)
	require.Nil(t, err)
	require.Equal(t, len(p2df), 1)

	var dir2df []schema.DataDir2DataFile
	rql = r.Table("datadir2datafile").Filter(r.Row.Field("datafile_id").Eq(newFile.ID))
	err = model.DirFiles.Qs(session).Rows(rql, &dir2df)
	require.Nil(t, err)
	require.Equal(t, len(dir2df), 1)

	deleteFile(newFile.ID)
}

func deleteFile(fileID string) {
	session := testutil.RSession()
	model.Files.Qs(session).Delete(fileID)

	rql := r.Table("project2datafile").Filter(r.Row.Field("datafile_id").Eq(fileID))
	var p2df []schema.Project2DataFile
	model.ProjectFiles.Qs(session).Rows(rql, &p2df)
	for _, entry := range p2df {
		model.ProjectFiles.Qs(session).Delete(entry.ID)
	}

	rql = r.Table("datadir2datafile").Filter(r.Row.Field("datafile_id").Eq(fileID))
	var d2df []schema.DataDir2DataFile
	model.DirFiles.Qs(session).Rows(rql, &d2df)
	for _, entry := range d2df {
		model.DirFiles.Qs(session).Delete(entry.ID)
	}
}
