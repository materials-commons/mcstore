package uploads

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

// chunks returns a list of the files in a given directory as a set of chunks. Chunks are
// returned in sorted order (ioutil.ReadDir will return the directory contents in sorted
// order).
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
