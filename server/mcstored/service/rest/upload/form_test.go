package upload

import (
	"bufio"
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultipart2FlowRequest(t *testing.T) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	fw := multipart.NewWriter(w)
	fw.WriteField("flowChunkNumber", "0")
	fw.WriteField("flowTotalChunks", "1")
	fw.WriteField("flowChunkSize", "5")
	fw.WriteField("flowTotalSize", "5")
	fw.WriteField("flowIdentifier", "unique")
	fw.WriteField("flowFileName", "test.txt")
	fw.WriteField("flowRelativePath", "test.txt")
	fw.WriteField("projectID", "project")
	fw.WriteField("directoryID", "directory")
	fw.WriteField("fileID", "file")
	fw.WriteField("chunkData", "hello")
	err := fw.Close()
	w.Flush() // Need to flush the writer.
	require.Nil(t, err, "Got unexpected error: %s", err)
	reader := bufio.NewReader(&b)
	req, err := multipart2FlowRequest(multipart.NewReader(reader, fw.Boundary()))
	require.Nil(t, err, "Got unexpected error: %s", err)
	require.NotNil(t, req)
	require.Equal(t, req.FlowChunkNumber, 0)
	require.Equal(t, req.FlowTotalChunks, 1)
	require.Equal(t, req.FlowChunkSize, 5)
	require.Equal(t, req.FlowTotalSize, 5)
	require.Equal(t, req.FlowIdentifier, "unique")
	require.Equal(t, req.FlowFileName, "test.txt")
	require.Equal(t, req.FlowRelativePath, "test.txt")
	require.Equal(t, req.ProjectID, "project")
	require.Equal(t, req.DirectoryID, "directory")
	require.Equal(t, req.FileID, "file")
	s := string(req.Chunk)
	require.Equal(t, s, "hello")
}
