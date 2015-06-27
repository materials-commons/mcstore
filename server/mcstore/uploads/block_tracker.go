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
	hasher       hash.Hash
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
	// requestBlockTracker is a shared instance of the block tracker. Since
	// blockTracker instances are synchronized this can be shared across
	// services and go routines.
	requestBlockTracker *blockTracker = newBlockTracker()
)

// newBlockTracker creates a new blockTracker instance.
func newBlockTracker() *blockTracker {
	return &blockTracker{
		reqBlocks: make(map[string]*blockTrackerEntry),
	}
}

func (self *blockTracker) idExists(id string) bool {
	var doesExist bool
	self.withReadLock(id, func(b *blockTrackerEntry) {
		doesExist = true
	})
	return doesExist
}

// setBlock marks a block as having the data written for it.
// The bitset starts counting at 0, but flowjs starts at 1
// so we adjust for the block in here.
func (self *blockTracker) setBlock(id string, block int) {
	self.withWriteLock(id, func(b *blockTrackerEntry) {
		b.bset.Set(uint(block - 1))
	})
}

// isBlockSet returns true if the block is already set.
func (self *blockTracker) isBlockSet(id string, block int) bool {
	var blockIsSet bool
	self.withReadLock(id, func(b *blockTrackerEntry) {
		blockIsSet = b.bset.Test(uint(block))
	})
	return blockIsSet
}

// load will load the blocks bitset for an id.
func (self *blockTracker) load(id string, numBlocks int) {
	self.withWriteLockNotExist(id, func() {
		bset := bitset.New(uint(numBlocks))
		self.reqBlocks[id] = &blockTrackerEntry{
			bset:   bset,
			hasher: md5.New(),
		}
	})
}

// clearBlock will unmark an block.
func (self *blockTracker) clearBlock(id string, block int) {
	self.withWriteLock(id, func(b *blockTrackerEntry) {
		b.bset.SetTo(uint(block-1), false)
	})
}

// markAllBlocks will mark all the blocks in the bitset
func (self *blockTracker) markAllBlocks(id string) {
	self.withWriteLock(id, func(b *blockTrackerEntry) {
		b.bset.ClearAll()
		b.bset = b.bset.Complement()
	})
}

// done returns true if all blocks have been marked for an id.
func (self *blockTracker) done(id string) bool {
	var allBlocksDone bool
	self.withReadLock(id, func(b *blockTrackerEntry) {
		allBlocksDone = b.bset.All()
	})
	return allBlocksDone
}

// clear removes an id from the block tracker.
func (self *blockTracker) clear(id string) {
	self.withWriteLock(id, func(b *blockTrackerEntry) {
		delete(self.reqBlocks, id)
	})
}

// hash will return the accumulated hash.
func (self *blockTracker) hash(id string) string {
	var hashStr string
	self.withWriteLock(id, func(b *blockTrackerEntry) {
		hashStr = fmt.Sprintf("%x", b.hasher.Sum(nil))
	})
	return hashStr
}

// addToHash will add to the hash for the blocks.
func (self *blockTracker) addToHash(id string, what []byte) {
	self.withWriteLock(id, func(b *blockTrackerEntry) {
		io.Copy(b.hasher, bytes.NewBuffer(what))
	})
}

// getBlocks returns a clone of the current bitset.
func (self *blockTracker) getBlocks(id string) *bitset.BitSet {
	var bset *bitset.BitSet
	self.withReadLock(id, func(b *blockTrackerEntry) {
		bset = b.bset.Clone()
	})
	return bset
}

// isExistingFile returns true if this entry represents a file
// that was previously loaded.
func (self *blockTracker) isExistingFile(id string) bool {
	var isExisting bool
	self.withReadLock(id, func(b *blockTrackerEntry) {
		isExisting = b.existingFile
	})
	return isExisting
}

// setIsExistingFile sets the entry as representing a file that
// was already uploaded.
func (self *blockTracker) setIsExistingFile(id string, existing bool) {
	self.withWriteLock(id, func(b *blockTrackerEntry) {
		b.existingFile = existing
	})
}

// withWriteLock will take out a write lock, look up the given id in the
// hash and call the given function with the lock if it finds an entry.
func (self *blockTracker) withWriteLock(id string, fn func(b *blockTrackerEntry)) {
	defer self.mutex.Unlock()
	self.mutex.Lock()
	if val, ok := self.reqBlocks[id]; ok {
		fn(val)
	} else {
		app.Log.Critf("withWriteLock critical error, unable to locate track id %s", id)
	}
}

// withWriteLockNotExist will take out a write lock, look up the given id in the hash
// and call the given function with the lock if it doesn't find an entry.
func (self *blockTracker) withWriteLockNotExist(id string, fn func()) {
	defer self.mutex.Unlock()
	self.mutex.Lock()
	if _, ok := self.reqBlocks[id]; !ok {
		fn()
	} else {
		app.Log.Critf("withWriteLockNotExist critical error, located track id %s", id)
	}
}

// withReadLock will take out a read lock, look up the given id in the
// hash and call the given function with the lock if it finds an entry.
func (self *blockTracker) withReadLock(id string, fn func(b *blockTrackerEntry)) {
	defer self.mutex.RUnlock()
	self.mutex.RLock()
	if val, ok := self.reqBlocks[id]; ok {
		fn(val)
	} else {
		app.Log.Critf("withReadLock critical error, unable to locate track id %s", id)
	}
}
