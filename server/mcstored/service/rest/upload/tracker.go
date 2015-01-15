package upload

import "sync"

type uploadTracker struct {
	mutex   sync.RWMutex
	tracker map[string]int32
}

func NewUploadTracker() *uploadTracker {
	return &uploadTracker{
		tracker: make(map[string]int32),
	}
}

// increment adds to the count of chunks, and returns the total count.
func (u *uploadTracker) increment(id string) int32 {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	val := u.tracker[id]
	val++
	u.tracker[id] = val
	return val
}

// count will return the count for a given id.
func (u *uploadTracker) count(id string) int32 {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	val := u.tracker[id]
	return val
}

// clear removes an id from the tracker.
func (u *uploadTracker) clear(id string) {
	defer u.mutex.Unlock()
	u.mutex.Lock()
	delete(u.tracker, id)
}
