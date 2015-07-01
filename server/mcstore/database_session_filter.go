package mcstore

import (
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
)

// databaseSessionFilter is a filter than creates new database sessions. It takes a
// function that creates new instances of the session.
type databaseSessionFilter struct {
	session func() (*r.Session, error)
}

// Filter will create a new database session and place it in the session request attribute. When control
// returns to the filter it will close the session.
func (f *databaseSessionFilter) Filter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if session, err := f.session(); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "Unable to connect to database")
	} else {
		request.SetAttribute("session", session)
		chain.ProcessFilter(request, response)
		session.Close()
	}
}
