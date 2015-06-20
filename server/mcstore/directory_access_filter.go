package mcstore

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
)

type directoryAccessFilter struct {
	projects dai.Projects
	dirs     dai.Dirs
	access   domain.Access
}

func newDirectoryAccessFilter(dirs dai.Dirs, projects dai.Projects, access domain.Access) *directoryAccessFilter {
	return &projectAccessFilter{
		projects: projects,
		access:   access,
		dirs:     dirs,
	}
}

type directoryIDAccess struct {
	DirectoryID string `json:"directory_id"`
}

func (f *directoryAccessFilter) Filter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	user := request.Attribute("user").(schema.User)
	project := request.Attribute("project").(schema.Project)
	var d directoryIDAccess

	if err := request.ReadEntity(&d); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, "No directory_id found")
		return
	}

	if directory, err := f.getDirectoryValidatingAccess(d.DirectoryID, project.ID, user.ID); err != nil {
		response.WriteErrorString(http.StatusUnauthorized, "No access to project")
	} else {
		request.SetAttribute("directory", *directory)
		chain.ProcessFilter(request, response)
	}
}

// getDirectoryValidatingAccess retrieves the directory with the given directoryID. It checks access to the
// directory and validates that the directory exists in the given project.
func (f *directoryAccessFilter) getDirectoryValidatingAccess(directoryID, projectID, user string) (*schema.Directory, error) {
	dir, err := f.dirs.ByID(directoryID)
	switch {
	case err != nil:
		return nil, err
	case !f.projects.HasDirectory(projectID, directoryID):
		return nil, app.ErrInvalid
	case !f.access.AllowedByOwner(projectID, user):
		return nil, app.ErrNoAccess
	default:
		return dir, nil
	}
}
