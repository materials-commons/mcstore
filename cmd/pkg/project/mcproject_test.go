package project

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var mcproject *MCProject

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	teardown()
	os.Exit(retCode)
}

func setup() {
	os.Mkdir(".mcproject", 0777)
}

func teardown() {
	os.RemoveAll(".mcproject")
}

func TestCreateDB(t *testing.T) {
	var err error
	mcproject, err = Create(".mcproject", "proj1", "proj1id")
	require.Nil(t, err, "Open failed: %s", err)
	require.NotNil(t, mcproject, "mcproject is nil")
	db := mcproject.db

	var projects []Project
	err = db.Select(&projects, "select * from project")
	require.Nil(t, err, "Select failed: %s", err)
	require.Equal(t, len(projects), 1, "Expected one project got %d", len(projects))
	proj := projects[0]
	require.Equal(t, proj.ProjectID, "proj1id", "Got wrong projectID: %s", proj.ProjectID)
	require.Equal(t, proj.Name, "proj1", "Got wrote name: %s", proj.Name)
}

func TestInsertDir(t *testing.T) {
	db := mcproject.db
	now := time.Now()
	dir := &Directory{
		DirectoryID: "abc123",
		Path:        "/tmp/dir",
		LastUpload:  now,
	}

	var err error
	dir, err = mcproject.InsertDirectory(dir)
	require.Nil(t, err, "insert failed %s", err)
	require.True(t, dir.ID != 0, "ID should not be 0: %#v", dir)

	var dirs []Directory
	err = db.Select(&dirs, "select * from directories")
	require.Nil(t, err, "Select failed: %s", err)
	require.Equal(t, 1, len(dirs), "Expected only 1 dir, got %d", len(dirs))
	require.Equal(t, "abc123", dirs[0].DirectoryID, "Got wrong directory id: %s", dirs[0].DirectoryID)
	require.True(t, dir.ID == dirs[0].ID, "Got unexpected id: %d", dirs[0].ID)
	require.True(t, dir.LastUpload == now, "Got unexpected last upload. Expected %#v, got %#v", dir.LastUpload, now)
}

func TestInsertFile(t *testing.T) {
	db := mcproject.db
	var dirs []Directory
	db.Select(&dirs, "select * from directories")
	f := &File{
		FileID:    "fileid123",
		Directory: dirs[0].ID,
		Name:      "test.txt",
		Size:      64 * 1024 * 1024 * 1024,
	}

	var err error
	f, err = mcproject.InsertFile(f)
	require.Nil(t, err, "insert failed: %s", err)
	require.True(t, f.ID != 0, "ID should not be 0: %#v", f)

	var files []File
	err = db.Select(&files, "select * from files")
	require.Nil(t, err, "Select failed: %s", err)
	require.Equal(t, len(files), 1, "Expected one file got %d", len(files))
	f0 := files[0]
	require.Equal(t, f0.FileID, "fileid123", "Got wrong fileid: %s", f0.FileID)
	require.Equal(t, f0.Name, "test.txt", "Got wrote name: %s", f0.Name)
	areEqual := f0.Size == (64 * 1024 * 1024 * 1024)
	require.True(t, areEqual, "Got wrong size: %d", f0.Size)
	require.True(t, f0.ID == f.ID, "Got unexpected id: %d", f0.ID)
}
