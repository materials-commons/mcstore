package upload

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"path/filepath"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

// form2FlowRequest reads a multipart upload form and converts it to a flow.Request.
func form2FlowRequest(request *restful.Request) (*flow.Request, error) {
	var (
		r      flow.Request
		err    error
		reader *multipart.Reader
		part   *multipart.Part
	)

	// Open multipart reading
	buf := new(bytes.Buffer)
	reader, err = request.Request.MultipartReader()
	if err != nil {
		return nil, err
	}

	// For each part identify its name to decide which field
	// to fill in the flow.Request.
	for {
		part, err = reader.NextPart()
		if err != nil {
			break
		}

		name := part.FormName()

		// Don't copy chunkData, it will be handled differently.
		if name != "chunkData" {
			io.Copy(buf, part)
		}

		switch name {
		case "flowChunkNumber":
			r.FlowChunkNumber = atoi32(buf.String())
		case "flowTotalChunks":
			r.FlowTotalChunks = atoi32(buf.String())
		case "flowChunkSize":
			r.FlowChunkSize = atoi32(buf.String())
		case "flowTotalSize":
			r.FlowTotalSize = atoi64(buf.String())
		case "flowIdentifier":
			r.FlowIdentifier = buf.String()
		case "flowFileName":
			r.FlowFileName = buf.String()
		case "flowRelativePath":
			r.FlowRelativePath = buf.String()
		case "projectID":
			r.ProjectID = buf.String()
		case "directoryID":
			r.DirectoryID = buf.String()
		case "fileID":
			r.FileID = buf.String()
		case "chunkData":
			// Get the chunk bytes.
			if r.Chunk, err = ioutil.ReadAll(part); err != nil {
				app.Log.Info(app.Logf("Error reading chunk, ReadAll returned %s", err))
			}
		}
		// Reset the buffer after each use.
		buf.Reset()
	}

	if err != io.EOF {
		return nil, err
	}

	return &r, nil
}

// atoi64 converts a string to an int64
func atoi64(str string) int64 {
	i, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		app.Log.Error(app.Logf("Error converting %s to an int", str))
		return -1
	}

	return i
}

// atoi32 converts a string to an int32
func atoi32(str string) int32 {
	i := atoi64(str)
	return int32(i)
}

// fileUploadPath creates the path to upload file chunks to.
func fileUploadPath(projectID, directoryID, fileID string) string {
	return filepath.Join(app.MCDir.Path(), "upload", projectID, directoryID, fileID)
}

// chunkPath creates the full path name for a chunk.
func chunkPath(uploadPath string, chunkNumber int32) string {
	n := fmt.Sprintf("%d", chunkNumber)
	return filepath.Join(uploadPath, n)
}
