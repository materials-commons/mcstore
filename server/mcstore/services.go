package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
)

// NewServicesContainer creates a new restful.Container made up of all
// the rest resources handled by the server.
func NewServicesContainer() *restful.Container {
	container := restful.NewContainer()

	databaseSessionFilter := &databaseSessionFilter{
		session: db.RSession,
	}
	container.Filter(databaseSessionFilter.Filter)

	apikeyFilter := newAPIKeyFilter()
	container.Filter(apikeyFilter.Filter)

	uploadResource := newUploadResource()
	container.Add(uploadResource.WebService())

	projectsResource := createProjectsResource()
	container.Add(projectsResource.WebService())

	return container
}

// projectsResource creates a new projects resource.
func createProjectsResource() rest.Service {
	return newProjectsResource(newDirService(), newProjectService())
}
