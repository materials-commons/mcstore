package dai

import (
	"testing"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
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

func TestInsert(t *testing.T) {
	//require.True(t, false, "Implementing cleanup")
	file := schema.NewFile("test1.txt", "test@mc.org")
	// Explicitly set the ID for our tests
	file.ID = "test1.txt"

	newFile, err := rfiles.Insert(&file, "test", "test")
	require.Nil(t, err)
	require.NotNil(t, newFile)
	require.Equal(t, newFile.ID, "test1.txt")

	deleteFile(file.ID)
}

func deleteFile(fileID string) {
	session := test.RSession()
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
