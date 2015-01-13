package upload

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type dirItem struct {
	os.FileInfo
	reader func() (io.Reader, error)
}

func fromDir(dir string) ([]Item, error) {
	finfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var dirItems []Item
	for _, finfo := range finfos {
		item := dirItem{
			FileInfo: finfo,
			reader: func() (io.Reader, error) {
				return os.Open(filepath.Join(dir, finfo.Name()))
			},
		}
		dirItems = append(dirItems, item)
	}
	return dirItems, err
}

func (d dirItem) Reader() (io.Reader, error) {
	return d.reader()
}

type byChunk []Item

func (l byChunk) Len() int      { return len(l) }
func (l byChunk) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byChunk) Less(i, j int) bool {
	iName, _ := strconv.Atoi(l[i].Name())
	jName, _ := strconv.Atoi(l[j].Name())
	return iName < jName
}
