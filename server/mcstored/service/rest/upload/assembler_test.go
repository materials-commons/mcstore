package upload

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssemblerWriteItemTo(t *testing.T) {
	var b bytes.Buffer
	destination := bufio.NewWriter(&b)
	item := testItem{
		name:    "name1",
		content: "content1",
	}

	writeItemTo(item, destination)
	require.Equal(t, "content1", b.String())
}

func TestAssemblerWriteEach(t *testing.T) {
	items := makeTestItems()
	var b bytes.Buffer
	destination := bufio.NewWriter(&b)
	writeEach(items, destination)
	require.Equal(t, "content2content1", b.String())
}

func TestAssemblerTo(t *testing.T) {
	// Test valid data case
	items := makeTestItems()
	var b bytes.Buffer
	destination := bufio.NewWriter(&b)
	finisher := &trackFinisher{}
	a := NewAssembler(items, finisher)
	err := a.To(destination)
	require.Nil(t, err, "Expected nil err, got %s", err)
	require.Equal(t, "content2content1", b.String())
	require.True(t, finisher.called)

	// Test failure case item couldn't create a reader
	finisher.called = false
	items = append(items, testItem{
		name:    "3",
		content: "content3",
		err:     errors.New("test item failing"),
	})
	a = NewAssembler(items, finisher)
	err = a.To(destination)
	require.NotNil(t, err, "Expected non nil err")
	require.False(t, finisher.called)

	// Test failure case finisher returned an error
	items = makeTestItems() // make items with no errors
	finisher.called = false
	finisher.err = errors.New("test finisher failing")
	a = NewAssembler(items, finisher)
	err = a.To(destination)
	require.NotNil(t, err, "Expected non nil err")
	require.True(t, finisher.called)

	// Test sorted assembly
	items = makeTestItems()
	finisher.called = false
	finisher.err = nil
	var b2 bytes.Buffer
	destination = bufio.NewWriter(&b2)
	sort.Sort(byChunk(items))
	a = NewAssembler(items, finisher)
	err = a.To(destination)
	require.Nil(t, err, "Expected nil err: %s", err)
	require.True(t, finisher.called)
	require.Equal(t, "content1content2", b2.String())
}

type testItem struct {
	name    string
	content string
	err     error
}

func (i testItem) Name() string {
	return i.name
}

func (i testItem) Reader() (io.Reader, error) {
	return ioutil.NopCloser(strings.NewReader(i.content)), i.err
}

func makeTestItems() []Item {
	testItems := []Item{
		testItem{
			name:    "2",
			content: "content2",
		},
		testItem{
			name:    "1",
			content: "content1",
		},
	}

	return testItems
}

type trackFinisher struct {
	called bool
	err    error
}

func (f *trackFinisher) Finish() error {
	f.called = true
	return f.err
}
