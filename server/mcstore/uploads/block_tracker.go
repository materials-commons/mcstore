package uploads

import (
	"sync"

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
}

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
