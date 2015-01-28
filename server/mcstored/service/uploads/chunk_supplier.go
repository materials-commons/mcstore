package uploads

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// A chunk contains a piece of the data to assemble.
type chunk interface {
	Name() string               // Name of the item
	Reader() (io.Reader, error) // Returns a reader to get at the items data
}

// A chunkSupplier supplies a list of chunks. Its useful for implementing
// different ways to get a list of chunks.
type chunkSupplier interface {
	chunks() ([]chunk, error)
}

// dirChunk implements the chunk interface. It provides an item for
// each file in a directory.
type dirChunk struct {
	os.FileInfo
	reader func() (io.Reader, error)
}

// Reader returns a new io.Reader for the given dirChunk file entry.
func (d dirChunk) Reader() (io.Reader, error) {
	return d.reader()
}

// A dirChunkSupplier returns a list of chunks from a directory.
type dirChunkSupplier struct {
	dir string
}

// newDirChunkSupplier creates a new dirChunkSupplier for the given directory path.
func newDirChunkSupplier(dir string) *dirChunkSupplier {
	return &dirChunkSupplier{
		dir: dir,
	}
}

// chunks returns a list of the files in a given directory as a set of chunks.
func (s *dirChunkSupplier) chunks() ([]chunk, error) {
	finfos, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	var dirChunks []chunk
	for _, finfo := range finfos {
		if !finfo.IsDir() {
			saveFinfo := finfo
			chunk := dirChunk{
				FileInfo: saveFinfo,
				reader: func() (io.Reader, error) {
					return os.Open(filepath.Join(s.dir, saveFinfo.Name()))
				},
			}
			dirChunks = append(dirChunks, chunk)
		}
	}
	return dirChunks, err
}

// byChunk provides sorting on chunk files. Chunk file names are numeric since
// chunks are numeric in ascending order.
type byChunk []chunk

func (l byChunk) Len() int      { return len(l) }
func (l byChunk) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byChunk) Less(i, j int) bool {
	iName, _ := strconv.Atoi(l[i].Name())
	jName, _ := strconv.Atoi(l[j].Name())
	return iName < jName
}
