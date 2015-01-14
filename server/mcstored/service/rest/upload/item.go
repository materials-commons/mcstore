package upload

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// A Item contains data to assemble.
type Item interface {
	Name() string               // Name of the item
	Reader() (io.Reader, error) // Returns a reader to get at the items data
}

// dirItem implements the Item interface. It provides an item for
// each file in a directory.
type dirItem struct {
	os.FileInfo
	reader func() (io.Reader, error)
}

// fromDir returns a list of the files in a given directory as a set of Items.
func fromDir(dir string) ([]Item, error) {
	finfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirItems []Item
	for _, finfo := range finfos {
		if !finfo.IsDir() {
			saveFinfo := finfo
			item := dirItem{
				FileInfo: saveFinfo,
				reader: func() (io.Reader, error) {
					return os.Open(filepath.Join(dir, saveFinfo.Name()))
				},
			}
			dirItems = append(dirItems, item)
		}
	}
	return dirItems, err
}

// Reader returns a new io.Reader for the given dirItem file entry.
func (d dirItem) Reader() (io.Reader, error) {
	return d.reader()
}

// byChunk provides sorting on chunk files. Chunk file names are numeric since
// chunks are numeric in ascending order.
type byChunk []Item

func (l byChunk) Len() int      { return len(l) }
func (l byChunk) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byChunk) Less(i, j int) bool {
	iName, _ := strconv.Atoi(l[i].Name())
	jName, _ := strconv.Atoi(l[j].Name())
	return iName < jName
}
