package app

import (
	"testing"

	"github.com/materials-commons/config"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	// Test MCDIR not set
	config.Set("MCDIR", "")
	panicFunc := func() {
		MCDir.Path()
	}
	require.Panics(t, panicFunc, "MCDIR not set and didn't panic")
	config.Set("MCDIR", "/tmp/mcdir")
	require.Equal(t, MCDir.Path(), "/tmp/mcdir", "Expected MCDir.Path to return /tmp/mcdir")
}

func TestFileDir(t *testing.T) {
	// Good file id
	fileID := "abc-defg-ghi-jkl-mnopqr"
	dir := MCDir.FileDir(fileID)
	require.Equal(t, "/tmp/mcdir/de/fg", dir, "Expected /tmp/mcdir/de/fg got %s", dir)

	// Bad file id
	fileID = "bad_file_id"
	dir = MCDir.FileDir(fileID)
	require.Equal(t, "", dir, "Expected '', got %s", dir)

	// Bad segment in file id
	fileID = "abc-def-ghi-jkl"
	dir = MCDir.FileDir(fileID)
	require.Equal(t, "", dir, "Expected '', got %s", dir)
}

func TestFilePath(t *testing.T) {
	// Good file id
	fileID := "abc-defg-ghi-jkl-mnopqr"
	path := MCDir.FilePath(fileID)
	require.Equal(t, "/tmp/mcdir/de/fg/abc-defg-ghi-jkl-mnopqr", path, "Expected /tmp/mcdir/de/fg/abc-defg-ghi-jkl-mnopqr, got %s", path)

	// Bad file id
	fileID = "bad_file_id"
	path = MCDir.FilePath(fileID)
	require.Equal(t, "", path, "Expected '', got %s", path)

	// Bad segment in file id
	fileID = "abc-def-ghi-jkl"
	path = MCDir.FilePath(fileID)
	require.Equal(t, "", path, "Expected '', got %s", path)
}

func TestFileConversionDir(t *testing.T) {
	// Good file id
	fileID := "abc-defg-ghi"
	path := MCDir.FileConversionDir(fileID)
	require.Equal(t, "/tmp/mcdir/de/fg/.conversion", path)

	// Bad file id
	fileID = "bad_file_id"
	path = MCDir.FileConversionDir(fileID)
	require.Equal(t, "", path, "Expected '', got %s", path)

	// Bad segment in file id
	fileID = "abc-def-ghi-jkl"
	path = MCDir.FileConversionDir(fileID)
	require.Equal(t, "", path, "Expected '', got %s", path)
}

func TestFilePathImageConversion(t *testing.T) {
	// Good file id
	fileID := "abc-defg-ghi"
	path := MCDir.FilePathImageConversion(fileID)
	require.Equal(t, "/tmp/mcdir/de/fg/.conversion/abc-defg-ghi.jpg", path, "Expected /tmp/mcdir/de/fg/.conversion/abc-defg-ghi.jpg, got %s", path)

	// Bad file id
	fileID = "bad_file_id"
	path = MCDir.FilePathImageConversion(fileID)
	require.Equal(t, "", path, "Expected '', got %s", path)

	// Bad segment in file id
	fileID = "abc-def-ghi-jkl"
	path = MCDir.FilePathImageConversion(fileID)
	require.Equal(t, "", path, "Expected '', got %s", path)
}
