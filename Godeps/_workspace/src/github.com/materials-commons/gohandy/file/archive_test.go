package file

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

var _ = fmt.Println

func TestUnpackTarGz(t *testing.T) {
	tdpath := filepath.Join("..", "test_data")
	p := filepath.Join(tdpath, "tartest.tar.gz")
	tr, err := NewTarGz(p)
	if err != nil {
		t.Fatalf("Failed to create TarReader %s\n", err.Error())
	}

	if err := tr.Unpack(tdpath); err != nil {
		t.Fatalf("Failed to unpack file %s\n", err.Error())
	}

	tpath := filepath.Join(tdpath, "t")
	checkContents(filepath.Join(tpath, "a"), "Hello a", 8, t)
	checkContents(filepath.Join(tpath, "b"), "Hello b", 8, t)
	checkContents(filepath.Join(tpath, "c"), "Hello c", 8, t)
}

func checkContents(fpath, expectedContents string, expectedLength int, t *testing.T) {
	contents, err := ioutil.ReadFile(fpath)

	if err != nil {
		t.Fatalf("Unable to read file %s\n", fpath)
	}

	if len(contents) != expectedLength {
		t.Fatalf("Expected length %d, got length %d\n", expectedLength, len(contents))
	}
	if strings.TrimSpace(string(contents)) != expectedContents {
		t.Fatalf("Unexpected file contents '%s'\n", string(contents))
	}

}
