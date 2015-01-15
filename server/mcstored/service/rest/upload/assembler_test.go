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

func TestAssemblerWriteItem(t *testing.T) {
	var b bytes.Buffer
	destination := bufio.NewWriter(&b)
	a := NewAssembler(nil, destination, nil)
	item := testItem{
		name:    "name1",
		content: "content1",
	}

	a.writeItem(item)
	require.Equal(t, "content1", b.String())
}

func TestAssemblerWriteEach(t *testing.T) {
	var b bytes.Buffer
	destination := bufio.NewWriter(&b)
	finisher := newTrackFinisher()
	a := NewAssembler(newTestItemSupplier(), destination, finisher)
	a.writeEach()
	require.Equal(t, "content2content1", b.String())
	require.False(t, finisher.called)
}

func TestAssemblerAssemble(t *testing.T) {
	// Test valid data case
	var b bytes.Buffer
	destination := bufio.NewWriter(&b)
	finisher := &trackFinisher{}
	a := NewAssembler(newTestItemSupplier(), destination, finisher)
	err := a.Assemble()
	require.Nil(t, err, "Expected nil err, got %s", err)
	require.Equal(t, "content2content1", b.String())
	require.True(t, finisher.called)

	// Test failure case item couldn't create a reader
	finisher.called = false
	items := makeTestItems()
	items = append(items, testItem{
		name:    "3",
		content: "content3",
		err:     errors.New("test item failing"),
	})
	a = NewAssembler(newTestItemSupplierFrom(items), destination, finisher)
	err = a.Assemble()
	require.NotNil(t, err, "Expected non nil err")
	require.False(t, finisher.called)

	// Test failure case finisher returned an error
	finisher.called = false
	finisher.err = errors.New("test finisher failing")
	a = NewAssembler(newTestItemSupplier(), destination, finisher)
	err = a.Assemble()
	require.NotNil(t, err, "Expected non nil err")
	require.True(t, finisher.called)

	// Test sorted assembly
	finisher.called = false
	finisher.err = nil
	var b2 bytes.Buffer
	destination = bufio.NewWriter(&b2)
	a = NewAssembler(newSortedTestItemSupplier(), destination, finisher)
	err = a.Assemble()
	require.Nil(t, err, "Expected nil err: %s", err)
	require.True(t, finisher.called)
	require.Equal(t, "content1content2", b2.String())
}

func TestMCDirAssemblerFactory(t *testing.T) {
	require.True(t, false, "Not implemented")
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

type testItemSupplier struct {
	items []Item
}

func newTestItemSupplier() *testItemSupplier {
	return &testItemSupplier{
		items: makeTestItems(),
	}
}

func newSortedTestItemSupplier() *testItemSupplier {
	items := makeTestItems()
	sort.Sort(byChunk(items))
	return &testItemSupplier{
		items: items,
	}
}

func newTestItemSupplierFrom(items []Item) *testItemSupplier {
	return &testItemSupplier{
		items: items,
	}
}

func (s *testItemSupplier) Items() ([]Item, error) {
	return s.items, nil
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

func newTrackFinisher() *trackFinisher {
	return &trackFinisher{}
}

func (f *trackFinisher) Finish() error {
	f.called = true
	return f.err
}
