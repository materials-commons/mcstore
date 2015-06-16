package uploads

import (
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"sync"

	"github.com/willf/bitset"
)

type requestBlockCount struct {
	blocksDone int
	numBlocks  int
	h          hash.Hash
}

// A uploadTracker tracks the block count for a given id.
// It synchronizes access so it can be safely used by
// multiple routines.
type blockCountTracker struct {
	mutex   sync.RWMutex
	tracker map[string]*requestBlockCount
}

// newBlockCountTracker creates a new uploadTracker.
func newBlockCountTracker() *blockCountTracker {
	return &blockCountTracker{
		tracker: make(map[string]*requestBlockCount),
	}
}

func (u *blockCountTracker) load(id string, numBlocks int) {
	defer u.mutex.Unlock()
	u.mutex.Lock()

	if _, ok := u.tracker[id]; !ok {
		req := &requestBlockCount{
			numBlocks: numBlocks,
			h:         md5.New(),
		}
		u.tracker[id] = req
	}
}

// increment adds to the count of chunks, and returns the total count.
func (u *blockCountTracker) setBlock(id string, block int) {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	req := u.tracker[id]
	req.blocksDone++
}

// count will return the count for a given id.
func (u *blockCountTracker) done(id string) bool {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	req := u.tracker[id]
	return req.blocksDone == req.numBlocks
}

func (u *blockCountTracker) clearBlock(id string, block int) {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	req := u.tracker[id]
	req.blocksDone--
}

// clear removes an id from the tracker.
func (u *blockCountTracker) clear(id string) {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	delete(u.tracker, id)
}

func (u *blockCountTracker) hash(id string) string {
	defer u.mutex.Unlock()
	u.mutex.Lock()

	h := u.tracker[id].h
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (u *blockCountTracker) addToHash(id string, what []byte) {
	h := u.tracker[id].h
	io.WriteString(h, string(what))
}

func (u *blockCountTracker) getBlocks(id string) *bitset.BitSet {
	return nil
}
