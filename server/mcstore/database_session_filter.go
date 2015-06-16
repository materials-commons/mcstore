package mcstore
import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db"
	"net/http"
)

type databaseSessionFilter struct{}

func (f *databaseSessionFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if session, err := db.RSession(); err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, "Unable to connect to database")
		return
	} else {
		req.SetAttribute("session", session)
		chain.ProcessFilter(req, resp)
		session.Close()
	}
}
