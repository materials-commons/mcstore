package uploads

import (
	"sync"

	"path/filepath"

	"hash"

	"crypto/md5"
	"fmt"
	"io"

	"github.com/materials-commons/gohandy/file"
	"github.com/willf/bitset"
)

type requestByBlockTracker struct {
	bset *bitset.BitSet
	h    hash.Hash
}

// blockTracker holds all the state of blocks for different upload requests.
// If a block has been successfully written then it is marked, otherwise
// the block is clear and no data has been written for it.
type blockTracker struct {
	mutex       sync.RWMutex
	requestPath requestPath
	fops        file.Operations
	reqBlocks   map[string]*requestByBlockTracker
}

// newBlockTracker creates a new blockTracker instance.
func newBlockTracker() *blockTracker {
	return &blockTracker{
		requestPath: &mcdirRequestPath{},
		fops:        file.OS,
		reqBlocks:   make(map[string]*requestByBlockTracker),
	}
}

// setBlock marks a block as having the data written for it.
func (t *blockTracker) setBlock(id string, block int) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	bset := t.reqBlocks[id].bset
	bset.Set(uint(block))
}

// load will load the blocks bitset for an id.
func (t *blockTracker) load(id string, numBlocks int) {
	defer t.mutex.Unlock()
	t.mutex.Lock()

	if _, ok := t.reqBlocks[id]; ok {
		return
	}

	bset := bitset.New(uint(numBlocks))
	t.reqBlocks[id] = &requestByBlockTracker{
		bset: bset,
		h:    md5.New(),
	}

	//	path := BlocksFile(t.requestPath, id)
	//	if f, err := t.fops.Open(path); err != nil {
	//		// File not found. Create new entry.
	//		bset := bitset.New(uint(numBlocks))
	//		t.reqBlocks[id] = bset
	//		t.writeBlocks(bset, id)
	//	} else {
	//		defer f.Close()
	//		var bset bitset.BitSet
	//		bset.ReadFrom(f)
	//		t.reqBlocks[id] = &bset
	//	}
}

//// persist writes the blocks bitset to the blocks file. It panics if it cannot
//// write the blocks file.
//func (t *blockTracker) persist(id string) {
//	defer t.mutex.Unlock()
//	t.mutex.Lock()
//	bset := t.reqBlocks[id]
//	t.writeBlocks(bset, id)
//}
//
//// persistAll writes all the blocks bitsets to their respective blocks file
//// for each id that is being tracked by the blocks tracker. It panics if
//// any of these fails to persist.
//func (t *blockTracker) persistAll() {
//	defer t.mutex.Unlock()
//	t.mutex.Lock()
//	for id, bset := range t.reqBlocks {
//		t.writeBlocks(bset, id)
//	}
//}
//
//// writeBlocks performs the operation of writing the blocks file. It doesn't
//// take out any locks and should never be called directly.
//func (t *blockTracker) writeBlocks(bset *bitset.BitSet, id string) {
//	path := BlocksFile(t.requestPath, id)
//	f, err := t.fops.Create(path)
//	if err != nil {
//		app.Log.Panicf("Can't write block file for request %s (path %s): %s", id, path, err)
//	}
//	defer f.Close()
//	bset.WriteTo(f)
//}

// clearBlock unmarks a block.
func (t *blockTracker) clearBlock(id string, block int) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	bset := t.reqBlocks[id].bset
	bset.SetTo(uint(block), false)
}

// done returns true if all blocks have been marked for an id.
func (t *blockTracker) done(id string) bool {
	defer t.mutex.Unlock()
	t.mutex.RLock()
	bset := t.reqBlocks[id].bset
	return bset.All()
}

// clear removes an id from the block tracker.
func (t *blockTracker) clear(id string) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	delete(t.reqBlocks, id)
}

func (t *blockTracker) hash(id string) string {
	defer t.mutex.Unlock()
	t.mutex.Lock()

	h := t.reqBlocks[id].h
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (t *blockTracker) addToHash(id string, what []byte) {
	h := t.reqBlocks[id].h
	io.WriteString(h, string(what))
}

// BlocksFile returns the path to the blocks file for a given id.
func BlocksFile(rpath requestPath, id string) (path string) {
	return filepath.Join(rpath.dirFromID(id), "blocks")
}
