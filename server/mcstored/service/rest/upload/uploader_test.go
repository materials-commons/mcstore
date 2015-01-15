package upload

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/stretchr/testify/require"
)

var ErrTestWriteFailure = errors.New("test write failure")

func TestUploaderProcessRequest(t *testing.T) {
	// Test no problems case
	rw := newTestRequestWriter()
	tracker := NewUploadTracker()
	u := NewUploader(rw, tracker)
	req1 := newReq("1", "hello", 2)
	req2 := newReq("1", "world", 2)
	err := u.processRequest(req1)
	require.Nil(t, err)
	err = u.processRequest(req2)
	require.Nil(t, err)
	rw.w.Flush()
	require.Equal(t, "helloworld", rw.b.String())

	// Test write returned error
	rw = newTestRequestWriter()
	rw.err = ErrTestWriteFailure
	tracker = NewUploadTracker()
	u = NewUploader(rw, tracker)
	err = u.processRequest(req1)
	require.Equal(t, err, ErrTestWriteFailure)
	require.Equal(t, 0, tracker.count("1"))

	// Test send all blocks, then send same block a second time.
	// Upload should be done, so tracker should not increment
	// the count.
	rw = newTestRequestWriter()
	tracker = NewUploadTracker()
	u = NewUploader(rw, tracker)
	err = u.processRequest(req1)
	require.Nil(t, err)
	require.Equal(t, 1, tracker.count("1"))
	err = u.processRequest(req2)
	require.Nil(t, err)
	require.Equal(t, 2, tracker.count("1"))

	err = u.processRequest(req2)
	require.Nil(t, err)
	require.Equal(t, 2, tracker.count("1"))
}

func TestUploaderAllBlocksUploaded(t *testing.T) {
	// Test nothing uploaded
	rw := newTestRequestWriter()
	tracker := NewUploadTracker()
	u := NewUploader(rw, tracker)
	req1 := newReq("1", "hello", 2)
	req2 := newReq("1", "world", 2)
	require.False(t, u.allBlocksUploaded(req1))
	require.False(t, u.allBlocksUploaded(req2))

	// Test one block uploaded
	err := u.processRequest(req1)
	require.Nil(t, err)
	require.False(t, u.allBlocksUploaded(req1))
	require.False(t, u.allBlocksUploaded(req2))

	// Test second (all) blocks uploaded
	err = u.processRequest(req1)
	require.Nil(t, err)
	require.True(t, u.allBlocksUploaded(req1))
	require.True(t, u.allBlocksUploaded(req2))
}

func newReq(uploadID, data string, totalChunks int32) *flow.Request {
	return &flow.Request{
		FlowIdentifier:  uploadID,
		Chunk:           []byte(data),
		FlowTotalChunks: totalChunks,
	}
}

type testRequestWriter struct {
	w   *bufio.Writer
	b   *bytes.Buffer
	err error
}

func newTestRequestWriter() *testRequestWriter {
	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)
	return &testRequestWriter{
		b: b,
		w: w,
	}
}

func (r *testRequestWriter) Write(req *flow.Request) error {
	if r.err != nil {
		return r.err
	}
	_, err := r.w.Write(req.Chunk)
	return err
}
