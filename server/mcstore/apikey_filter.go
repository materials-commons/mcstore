package mcstore

import (
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// apikeyFilter implements a filter for checking the apikey
// passed in with a request.
type apikeyFilter struct {
	keycache *apikeyCache
}

// newAPIKeyFilter creates a new apikeyFilter instance.
func newAPIKeyFilter(keycache *apikeyCache) *apikeyFilter {
	return &apikeyFilter{
		keycache: keycache,
	}
}

// changes will monitor for changes to user apikeys and will
// update the server with the new key.
//func (f *apikeyFilter) changes() {
//	var session *r.Session
//	go func() {
//		var c struct {
//			NewUserValue schema.User `gorethink:"new_value"`
//			OldUserValue schema.User `gorethink:"old_value"`
//		}
//		users, _ := r.Table("users").Changes().Run(session)
//		for users.Next(&c) {
//			switch {
//			case c.OldUserValue.ID == "":
//				// no old id, so new user added
//				f.keycache.addKey(c.NewUserValue.APIKey, &c.NewUserValue)
//			case c.OldUserValue.APIKey != "" && c.OldUserValue.APIKey != c.NewUserValue.APIKey:
//				f.keycache.resetKey(c.OldUserValue.APIKey, c.NewUserValue.APIKey, &c.NewUserValue)
//			}
//		}
//	}()
//}

// Filter implements the Filter interface for apikey lookup. It checks if an apikey is
// valid. If the apikey is found it sets the "user" attribute to the user structure. If
// the apikey is invalid then the filter doesn't pass the request on, and instead returns
// an http.StatusUnauthorized.
func (f *apikeyFilter) Filter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	apikey := request.Request.URL.Query().Get("apikey")
	session := request.Attribute("session").(*r.Session)
	rusers := dai.NewRUsers(session)
	user, found := f.getUser(apikey, rusers)
	if !found {
		response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
	} else {
		request.SetAttribute("user", *user)
		chain.ProcessFilter(request, response)
	}
}

// getUser matches the user with the apikey. If it cannot find a match then it returns false.
// getUser caches the key/user pair in f.apikeys.
func (f *apikeyFilter) getUser(apikey string, users dai.Users) (*schema.User, bool) {
	if apikey == "" {
		// No key was passed.
		return nil, false
	}

	if user := f.keycache.getUser(apikey); user == nil {
		user, err := users.ByAPIKey(apikey)
		if err != nil {
			return nil, false
		}
		f.keycache.addKey(apikey, user)
		return user, true
	} else {
		return user, true
	}
}
