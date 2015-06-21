package mcstore

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type directoryFilter struct {
	projects dai.Projects
	dirs     dai.Dirs
}

func newDirectoryAccessFilter(dirs dai.Dirs, projects dai.Projects) *directoryFilter {
	return &directoryFilter{
		projects: projects,
		dirs:     dirs,
	}
}

type directoryIDAccess struct {
	DirectoryID string `json:"directory_id"`
}

func (f *directoryFilter) Filter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	project := request.Attribute("project").(schema.Project)
	var d directoryIDAccess

	if err := request.ReadEntity(&d); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, "No directory_id found")
		return
	}

	if directory, err := f.getDirectory(d.DirectoryID, project.ID); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, "Unknown directory_id")
	} else {
		request.SetAttribute("directory", *directory)
		chain.ProcessFilter(request, response)
	}
}

// getDirectoryValidatingAccess retrieves the directory with the given directoryID. It checks access to the
// directory and validates that the directory exists in the given project.
func (f *directoryFilter) getDirectory(directoryID, projectID string) (*schema.Directory, error) {
	dir, err := f.dirs.ByID(directoryID)
	switch {
	case err != nil:
		return nil, err
	case !f.projects.HasDirectory(projectID, directoryID):
		return nil, app.ErrInvalid
	default:
		return dir, nil
	}
}
