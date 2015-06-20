package uploads

import (
	"sync"

	"hash"

	"crypto/md5"
	"fmt"

	"bytes"
	"io"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/willf/bitset"
)

type blockTrackerEntry struct {
	bset         *bitset.BitSet
	h            hash.Hash
	existingFile bool
}

// blockTracker holds all the state of blocks for different upload requests.
// If a block has been successfully written then it is marked, otherwise
// the block is clear and no data has been written for it.
type blockTracker struct {
	mutex     sync.RWMutex
	reqBlocks map[string]*blockTrackerEntry
}

var (
	requestBlockTracker *blockTracker = newBlockTracker()
)

// newBlockTracker creates a new blockTracker instance.
func newBlockTracker() *blockTracker {
	return &blockTracker{
		reqBlocks: make(map[string]*blockTrackerEntry),
	}
}

func (t *blockTracker) idExists(id string) bool {
	var doesExist bool
	t.withReadLock(id, func(b *blockTrackerEntry) {
		doesExist = true
	})
	return doesExist
}

// setBlock marks a block as having the data written for it.
// The bitset starts counting at 0, but flowjs starts at 1
// so we adjust for the block in here.
func (t *blockTracker) setBlock(id string, block int) {
	t.withWriteLock(id, func(b *blockTrackerEntry) {
		bset := t.reqBlocks[id].bset
		bset.Set(uint(block - 1))
	})
}

// isBlockSet returns true if the block is already set.
func (t *blockTracker) isBlockSet(id string, block int) bool {
	var blockIsSet bool
	t.withReadLock(id, func(b *blockTrackerEntry) {
		bset := t.reqBlocks[id].bset
		blockIsSet = bset.Test(uint(block))
	})
	return blockIsSet
}

// load will load the blocks bitset for an id.
func (t *blockTracker) load(id string, numBlocks int) {
	t.withWriteLockNotExist(id, func(b *blockTrackerEntry) {
		bset := bitset.New(uint(numBlocks))
		t.reqBlocks[id] = &blockTrackerEntry{
			bset: bset,
			h:    md5.New(),
		}
	})
}

// clearBlock will unmark an block.
func (t *blockTracker) clearBlock(id string, block int) {
	t.withWriteLock(id, func(b *blockTrackerEntry) {
		bset := t.reqBlocks[id].bset
		bset.SetTo(uint(block-1), false)
	})
}

// markAllBlocks will mark all the blocks in the bitset
func (t *blockTracker) markAllBlocks(id string) {
	t.withWriteLock(id, func(b *blockTrackerEntry) {
		bset := t.reqBlocks[id].bset
		bset.ClearAll()
		bset = bset.Complement()
	})
}

// done returns true if all blocks have been marked for an id.
func (t *blockTracker) done(id string) bool {
	var allBlocksDone bool
	t.withReadLock(id, func(b *blockTrackerEntry) {
		allBlocksDone = b.bset.All()
	})
	return allBlocksDone
}

// clear removes an id from the block tracker.
func (t *blockTracker) clear(id string) {
	t.withWriteLock(id, func(b *blockTrackerEntry) {
		delete(t.reqBlocks, id)
	})
}

// hash will return the accumulated hash.
func (t *blockTracker) hash(id string) string {
	var hashStr string
	t.withWriteLock(id, func(b *blockTrackerEntry) {
		h := t.reqBlocks[id].h
		hashStr = fmt.Sprintf("%x", h.Sum(nil))
	})
	return hashStr
}

// addToHash will add to the hash for the blocks.
func (t *blockTracker) addToHash(id string, what []byte) {
	t.withWriteLock(id, func(b *blockTrackerEntry) {
		h := t.reqBlocks[id].h
		io.Copy(h, bytes.NewBuffer(what))
	})
}

// getBlocks returns a clone of the current bitset.
func (t *blockTracker) getBlocks(id string) *bitset.BitSet {
	var bset *bitset.BitSet
	t.withReadLock(id, func(b *blockTrackerEntry) {
		bset = b.bset.Clone()
	})
	return bset
}

// isExistingFile returns true if this entry represents a file
// that was previously loaded.
func (t *blockTracker) isExistingFile(id string) bool {
	var isExisting bool
	t.withReadLock(id, func(b *blockTrackerEntry) {
		isExisting = b.existingFile
	})
	return isExisting
}

// setIsExistingFile sets the entry as representing a file that
// was already uploaded.
func (t *blockTracker) setIsExistingFile(id string, existing bool) {
	t.withWriteLock(id, func(b *blockTrackerEntry) {
		b.existingFile = existing
	})
}

// withWriteLock will take out a write lock, look up the given id in the
// hash and call the given function with the lock if it finds an entry.
func (t *blockTracker) withWriteLock(id string, fn func(b *blockTrackerEntry)) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	if val, ok := t.reqBlocks[id]; ok {
		fn(val)
	} else {
		app.Log.Critf("withWriteLock critical error, unable to locate track id %s", id)
	}
}

// withWriteLockNotExist will take out a write lock, look up the given id in the hash
// and call the given function with the lock if it doesn't find an entry.
func (t *blockTracker) withWriteLockNotExist(id string, fn func(b *blockTrackerEntry)) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	if val, ok := t.reqBlocks[id]; !ok {
		fn(val)
	} else {
		app.Log.Critf("withWriteLockNotExist critical error, located track id %s", id)
	}
}

// withReadLock will take out a read lock, look up the given id in the
// hash and call the given function with the lock if it finds an entry.
func (t *blockTracker) withReadLock(id string, fn func(b *blockTrackerEntry)) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	if val, ok := t.reqBlocks[id]; ok {
		fn(val)
	} else {
		app.Log.Critf("withReadLock critical error, unable to locate track id %s", id)
	}
}
