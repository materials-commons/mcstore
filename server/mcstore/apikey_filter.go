package mcstore

import (
	"net/http"

	"sync"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// apikeyFilter implements a filter for checking the apikey
// passed in with a request.
type apikeyFilter struct {
	users   dai.Users
	apikeys map[string]*schema.User
	mutex   sync.RWMutex
}

// newAPIKeyFilter creates a new apikeyFilter instance.
func newAPIKeyFilter(users dai.Users) *apikeyFilter {
	return &apikeyFilter{
		users:   users,
		apikeys: make(map[string]*schema.User),
	}
}

// Filter implements the Filter interface for apikey lookup. It checks if an apikey is
// valid. If the apikey is found it sets the "user" attribute to the user structure. If
// the apikey is invalid then the filter doesn't pass the request on, and instead returns
// an http.StatusUnauthorized.
func (f *apikeyFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	apikey := req.Request.URL.Query().Get("apikey")
	user, found := f.getUser(apikey)
	if !found {
		resp.WriteErrorString(http.StatusUnauthorized, "Not authorized")
		return
	}

	req.SetAttribute("user", *user)
	chain.ProcessFilter(req, resp)
}

// getUser matches the user with the apikey. If it cannot find
// a match then it returns false. getUser caches the key/user
// pair in f.apikeys.
func (f *apikeyFilter) getUser(apikey string) (*schema.User, bool) {
	if apikey == "" {
		// No key was passed.
		return nil, false
	}

	user, found := f.getUserWithReadLock(apikey)
	if !found {
		user, err := f.users.ByAPIKey(apikey)
		if err != nil {
			return nil, false
		}
		f.setUserWithWriteLock(apikey, user)
		return user, true
	}

	return user, true
}

// getUserWithReadLock will acquire a read lock and look the user up in the
// hash table cache.
func (f *apikeyFilter) getUserWithReadLock(apikey string) (*schema.User, bool) {
	defer f.mutex.RUnlock()
	f.mutex.RLock()

	user, found := f.apikeys[apikey]
	return user, found
}

// setUserWithWriteLock will acquire a write lock and add the user to the
// hash table cache.
func (f *apikeyFilter) setUserWithWriteLock(apikey string, user *schema.User) {
	defer f.mutex.Unlock()
	f.mutex.Lock()

	f.apikeys[apikey] = user
}
