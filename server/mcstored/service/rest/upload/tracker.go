package upload

import "sync"

type chunkTracker struct {
	mutex   sync.RWMutex
	tracker map[string]int32
}

func newChunkTracker() *chunkTracker {
	return &chunkTracker{
		tracker: make(map[string]int32),
	}
}

// addChunk adds to the count of chunks, and returns the total count.
func (c *chunkTracker) addChunk(id string) int32 {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	val := c.tracker[id]
	val++
	c.tracker[id] = val
	return val
}

// clear removes an id from the tracker.
func (c *chunkTracker) clear(id string) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	delete(c.tracker, id)
}
