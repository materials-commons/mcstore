package mcstore

import (
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type directoryFilterDAI struct {
	projects dai.Projects
	dirs     dai.Dirs
}

func newDirectoryFilterDAI(session *r.Session) *directoryFilterDAI {
	return &directoryFilterDAI{
		projects: dai.NewRProjects(session),
		dirs:     dai.NewRDirs(session),
	}
}

type directoryIDAccess struct {
	DirectoryID string `json:"directory_id"`
}

func directoryFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	project := request.Attribute("project").(schema.Project)
	session := request.Attribute("session").(*r.Session)
	var d directoryIDAccess

	if err := request.ReadEntity(&d); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, "No directory_id found")
		return
	}

	filterDAI := newDirectoryFilterDAI(session)
	if directory, err := filterDAI.getDirectory(d.DirectoryID, project.ID); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, "Unknown directory_id")
	} else {
		request.SetAttribute("directory", *directory)
		chain.ProcessFilter(request, response)
	}
}

// getDirectoryValidatingAccess retrieves the directory with the given directoryID. It checks
// that the directory exists in the given project.
func (d *directoryFilterDAI) getDirectory(directoryID, projectID string) (*schema.Directory, error) {
	dir, err := d.dirs.ByID(directoryID)
	switch {
	case err != nil:
		return nil, err
	case !d.projects.HasDirectory(projectID, directoryID):
		return nil, app.ErrInvalid
	default:
		return dir, nil
	}
}
