package mcstore

import (
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
)

type databaseSessionFilter struct {
	session func() (*r.Session, error)
}

func (f *databaseSessionFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if session, err := f.session(); err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, "Unable to connect to database")
	} else {
		req.SetAttribute("session", session)
		chain.ProcessFilter(req, resp)
		session.Close()
	}
}
