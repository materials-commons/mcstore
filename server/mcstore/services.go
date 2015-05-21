package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/uploads"
)

// NewServicesContainer creates a new restful.Container made up of all
// the rest resources handled by the server.
func NewServicesContainer() *restful.Container {
	container := restful.NewContainer()

	apikeyFilter := NewAPIKeyFilter(dai.NewRUsers(db.RSessionMust()))
	container.Filter(apikeyFilter.Filter)

	uploadResource := createUploadsResource()
	container.Add(uploadResource.WebService())

	projectsResource := createProjectsResource()
	container.Add(projectsResource.WebService())

	return container
}

// uploadResource creates a new upload resource.
func createUploadsResource() rest.Service {
	return newUploadResource(uploads.NewUploadService(), uploads.NewIDService(), NewDirService())
}

// projectsResource creates a new projects resource.
func createProjectsResource() rest.Service {
	return newProjectsResource(NewDirService(), NewProjectService())
}
