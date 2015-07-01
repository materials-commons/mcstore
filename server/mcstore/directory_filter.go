package mcstore

import (
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws"
)

func directoryFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	project := request.Attribute("project").(schema.Project)
	session := request.Attribute("session").(*r.Session)

	var d struct {
		DirectoryID string `json:"directory_id"`
	}

	err := request.ReadEntity(&d)

	switch {
	case err != nil:
		response.WriteErrorString(http.StatusNotAcceptable, "No directory_id found")
	case d.DirectoryID == "":
		response.WriteErrorString(http.StatusNotAcceptable, "No directory_id found")
	default:
		dirs := dai.NewRDirs(session)
		projects := dai.NewRProjects(session)
		if dir, err := dirs.ByID(d.DirectoryID); err != nil {
			ws.WriteError(err, response)
		} else if !projects.HasDirectory(project.ID, dir.ID) {
			response.WriteErrorString(http.StatusNotAcceptable, "Uknown directory for project")
		} else {
			request.SetAttribute("directory", *dir)
			chain.ProcessFilter(request, response)
		}
	}
}
