package mcstore

import (
	"sync"

	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// apikeyCache is a mutex protected cache of users
// mapped to apikeys.
type apikeyCache struct {
	mutex   *sync.RWMutex
	apikeys map[string]*schema.User
}

// apiKeyCache is a single reference to a cache that
// can be shared by modules and go routines.
var apiKeyCache *apikeyCache = newAPIKeyCache()

// newAPIKeyCache creates a new apikeyCache.
func newAPIKeyCache() *apikeyCache {
	return &apikeyCache{
		apikeys: make(map[string]*schema.User),
		mutex:   &sync.RWMutex{},
	}
}

// getUser returns a user matching the given key. It
// returns nil if no user matches the key.
func (c *apikeyCache) getUser(apikey string) *schema.User {
	var user *schema.User
	c.withReadLock(apikey, func(u *schema.User) {
		user = u
	})
	return user
}

// addKey will add a new apikey/user mapping. If there is
// already an entry matching this key then nothing happens.
func (c *apikeyCache) addKey(apikey string, user *schema.User) {
	c.withWriteLockNotExisting(apikey, func() {
		c.apikeys[apikey] = user
	})
}

// resetKey deletes the oldkey and adds the newkey pointing to user. If
// oldkey doesn't exist nothing happens.
func (c *apikeyCache) resetKey(oldkey, newkey string, user *schema.User) {
	c.withWriteLock(oldkey, func() {
		delete(c.apikeys, oldkey)
		c.apikeys[newkey] = user
	})
}

// withReadLock should only be called by the apikeyCache. It takes out
// a read lock and if the given key is found it calls the passed func.
func (c *apikeyCache) withReadLock(id string, fn func(user *schema.User)) {
	defer c.mutex.RUnlock()
	c.mutex.RLock()
	if user, found := c.apikeys[id]; found {
		fn(user)
	}
}

// withWriteLock should only be called by the apikeyCache. It takes out
// a write lock and if the given key is found calls func.
func (c *apikeyCache) withWriteLock(id string, fn func()) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if _, found := c.apikeys[id]; found {
		fn()
	}
}

// withWriteLockNotExisting should only be called by the apikeyCache. It
// takes out a write lock and if the given key is not found calls func.
func (c *apikeyCache) withWriteLockNotExisting(id string, fn func()) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if _, found := c.apikeys[id]; !found {
		fn()
	}
}
