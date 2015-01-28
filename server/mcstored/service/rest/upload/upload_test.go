package upload

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/stretchr/testify/require"
)

func TestUploadServiceCompleteUpload(t *testing.T) {
	rw := newSeparateItemRequestWriter()
	tracker := NewUploadTracker()
	uploader := NewUploader(rw, tracker)
	var b bytes.Buffer
	dest := bufio.NewWriter(&b)
	finisher := newTrackFinisher()
	af := newSeparateItemAssemblerFactory(rw, dest, finisher)
	uploadResource := NewResource(uploader, af, nil)
	container := restful.NewContainer()
	container.Add(uploadResource.WebService())
	ts := httptest.NewServer(container)
	defer ts.Close()

	surl := ts.URL + "/upload/chunk"
	// Send data chunks
	req1 := newFlowReq1()
	req2 := newFlowReq2()
	req1b, ctype1 := flowRequest2Form(req1)
	req2b, ctype2 := flowRequest2Form(req2)

	resp, err := http.Post(surl, ctype1, req1b)
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)

	resp, err = http.Post(surl, ctype2, req2b)
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)

	dest.Flush() // Flush dest to get bytes into b.

	require.True(t, af.called)
	require.Equal(t, "helloworlds", b.String())
	require.True(t, finisher.called)
}

func TestUploadServiceIncompleteUpload(t *testing.T) {
	rw := newSeparateItemRequestWriter()
	tracker := NewUploadTracker()
	uploader := NewUploader(rw, tracker)
	var b bytes.Buffer
	dest := bufio.NewWriter(&b)
	finisher := newTrackFinisher()
	af := newSeparateItemAssemblerFactory(rw, dest, finisher)
	uploadResource := NewResource(uploader, af, nil)
	container := restful.NewContainer()
	container.Add(uploadResource.WebService())
	ts := httptest.NewServer(container)
	defer ts.Close()

	surl := ts.URL + "/upload/chunk"
	// Send data chunks
	req1 := newFlowReq1()

	req1b, ctype1 := flowRequest2Form(req1)

	resp, err := http.Post(surl, ctype1, req1b)
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)

	dest.Flush() // Flush dest to get bytes into b

	// No assembly should have been performed
	require.False(t, af.called)
	require.Equal(t, "", b.String())

	// Which means the finisher wasn't called either
	require.False(t, finisher.called)

	require.Equal(t, 1, len(rw.items))
	require.Equal(t, "hello", rw.items[0].String())
}

func TestUploadServiceErrorWritingUpload(t *testing.T) {
	rw := newSeparateItemRequestWriter()
	rw.err = ErrTestWriteFailure
	tracker := NewUploadTracker()
	uploader := NewUploader(rw, tracker)
	var b bytes.Buffer
	dest := bufio.NewWriter(&b)
	finisher := newTrackFinisher()
	af := newSeparateItemAssemblerFactory(rw, dest, finisher)
	uploadResource := NewResource(uploader, af, nil)
	container := restful.NewContainer()
	container.Add(uploadResource.WebService())
	ts := httptest.NewServer(container)
	defer ts.Close()

	surl := ts.URL + "/upload/chunk"
	// Send data chunks
	req1 := newFlowReq1()

	req1b, ctype1 := flowRequest2Form(req1)

	resp, err := http.Post(surl, ctype1, req1b)
	require.Nil(t, err)
	require.Equal(t, 500, resp.StatusCode)
	dest.Flush() // Flush dest to get bytes into b

	require.Equal(t, 0, len(rw.items))
	require.Equal(t, "", b.String())
	require.False(t, af.called)
	require.False(t, finisher.called)
}

// func TestMD5(t *testing.T) {
// 	checksum, err := file.HashStr(md5.New(), "/home/gtarcea/workspace/src/github.com/materials-commons/mcstore/server/mcstored/service/rest/upload/hashit")
// 	fmt.Println(err)
// 	fmt.Println(checksum)
// }

func newFlowReq1() *flow.Request {
	return &flow.Request{
		FlowChunkNumber:  1,
		FlowTotalChunks:  2,
		FlowChunkSize:    5,
		FlowTotalSize:    11,
		FlowIdentifier:   "unique",
		FlowFileName:     "test.txt",
		FlowRelativePath: "test.txt",
		ProjectID:        "project",
		DirectoryID:      "directory",
		FileID:           "file",
		Chunk:            []byte("hello"),
	}
}

func newFlowReq2() *flow.Request {
	req := newFlowReq1()
	req.FlowChunkSize = 6
	req.FlowChunkNumber = 2
	req.Chunk = []byte("worlds")
	return req
}

type separateItemRequestWriter struct {
	items []*bytes.Buffer
	err   error
}

func newSeparateItemRequestWriter() *separateItemRequestWriter {
	return &separateItemRequestWriter{}
}

func (r *separateItemRequestWriter) Write(req *flow.Request) error {
	if r.err != nil {
		return r.err
	}

	if len(r.items) >= int(req.FlowChunkNumber) {
		// Don't write chunk number twice
		return nil
	}

	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)
	_, err := w.Write(req.Chunk)
	w.Flush()
	r.items = append(r.items, b)
	return err
}

type bufferItem struct {
	b     *bytes.Buffer
	index int
}

func newBufferItem(b *bytes.Buffer, index int) bufferItem {
	return bufferItem{
		b:     b,
		index: index,
	}
}

func (i bufferItem) Name() string {
	return string(i.index)
}

func (i bufferItem) Reader() (io.Reader, error) {
	return i.b, nil
}

type separateItemSupplier struct {
	// keep a reference to the separateItemRequestWriter so we can
	// get at its list of buffers and turn them into Items
	rw *separateItemRequestWriter
}

func newSeparateItemSupplier(rw *separateItemRequestWriter) *separateItemSupplier {
	return &separateItemSupplier{
		rw: rw,
	}
}

func (s *separateItemSupplier) Items() ([]Item, error) {
	var items []Item
	for index, buf := range s.rw.items {
		bitem := newBufferItem(buf, index)
		items = append(items, bitem)
	}
	return items, nil
}

type separateItemAssemblerFactory struct {
	// keep a reference to the separateItemRequestWriter so we can
	// get at its list of buffers and turn them into Items
	rw       *separateItemRequestWriter
	dest     io.Writer
	finisher Finisher
	called   bool
}

func newSeparateItemAssemblerFactory(rw *separateItemRequestWriter, dest io.Writer, finisher Finisher) *separateItemAssemblerFactory {
	return &separateItemAssemblerFactory{
		rw:       rw,
		dest:     dest,
		finisher: finisher,
	}
}

func (f *separateItemAssemblerFactory) Assembler(req *flow.Request, owner string) *Assembler {
	f.called = true
	itemSupplier := newSeparateItemSupplier(f.rw)
	return NewAssembler(itemSupplier, f.dest, f.finisher)
}
