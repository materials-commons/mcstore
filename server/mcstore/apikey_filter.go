package mcstore

import (
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
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

// Filter implements the Filter interface for apikey lookup. It checks if an apikey is
// valid. If the apikey is found it sets the "user" attribute to the user structure. If
// the apikey is invalid then the filter doesn't pass the request on, and instead returns
// an http.StatusUnauthorized.
func (f *apikeyFilter) Filter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if apikey := request.Request.URL.Query().Get("apikey"); apikey == "" {
		// No or blank apikey passed in
		response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
	} else {
		session := request.Attribute("session").(*r.Session)
		rusers := dai.NewRUsers(session)
		if user := f.getUser(apikey, rusers); user == nil {
			response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
		} else {
			request.SetAttribute("user", *user)
			chain.ProcessFilter(request, response)
		}
	}
}

// getUser matches the user with the apikey. If it cannot find a match then it returns false.
func (f *apikeyFilter) getUser(apikey string, users dai.Users) *schema.User {
	if user := f.keycache.getUser(apikey); user != nil {
		return user
	}
	return f.loadUserFromDB(apikey, users)
}

// loadUserFromDB will look up the user by the apikey. If found it will add the
// user to cache and return the user. Otherwise it will return nil.
func (f *apikeyFilter) loadUserFromDB(apikey string, users dai.Users) *schema.User {
	if user, err := users.ByAPIKey(apikey); err != nil {
		app.Log.Infof("Look up user by apikey failed: %s\n", err)
		return nil
	} else {
		f.keycache.addKey(apikey, user)
		return user
	}
}
