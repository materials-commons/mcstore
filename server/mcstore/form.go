package mcstore

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strconv"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

// form2FlowRequest reads a multipart upload form and converts it to a flow.Request.
func form2FlowRequest(request *restful.Request) (*flow.Request, error) {
	reader, err := request.Request.MultipartReader()
	if err != nil {
		return nil, err
	}
	return multipart2FlowRequest(reader)
}

// multipart2FlowRequest creates a new flow.Request from the multipart Reader.
// If returns an error if the form is invalid.
func multipart2FlowRequest(reader *multipart.Reader) (*flow.Request, error) {
	var (
		r    flow.Request
		err  error
		part *multipart.Part
	)

	// Open multipart reading
	buf := new(bytes.Buffer)

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
				app.Log.Infof("Error reading chunk, ReadAll returned %s", err)
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
		app.Log.Errorf("Error converting %s to an int", str)
		return -1
	}

	return i
}

// atoi32 converts a string to an int32
func atoi32(str string) int32 {
	i := atoi64(str)
	return int32(i)
}

// flowRequest2Form creates a form bytes buffer from a flow.Request.
func flowRequest2Form(req *flow.Request) (*bytes.Buffer, string) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	fw := multipart.NewWriter(w)
	fw.WriteField("flowChunkNumber", strconv.Itoa(int(req.FlowChunkNumber)))
	fw.WriteField("flowTotalChunks", strconv.Itoa(int(req.FlowTotalChunks)))
	fw.WriteField("flowChunkSize", strconv.Itoa(int(req.FlowChunkSize)))
	fw.WriteField("flowTotalSize", strconv.Itoa(int(req.FlowTotalSize)))
	fw.WriteField("flowIdentifier", req.FlowIdentifier)
	fw.WriteField("flowFileName", req.FlowFileName)
	fw.WriteField("flowRelativePath", req.FlowRelativePath)
	fw.WriteField("projectID", req.ProjectID)
	fw.WriteField("directoryID", req.DirectoryID)
	fw.WriteField("fileID", req.FileID)
	fw.WriteField("chunkData", string(req.Chunk))
	contentType := fw.FormDataContentType()
	fw.Close()
	w.Flush() // Need to flush the writer
	return &b, contentType
}
