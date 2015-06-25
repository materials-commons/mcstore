package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db"
)

// NewServicesContainer creates a new restful.Container made up of all
// the rest resources handled by the server.
func NewServicesContainer() *restful.Container {
	container := restful.NewContainer()

	databaseSessionFilter := &databaseSessionFilter{
		session: db.RSession,
	}
	container.Filter(databaseSessionFilter.Filter)

	apikeyFilter := newAPIKeyFilter(apiKeyCache)
	container.Filter(apikeyFilter.Filter)

	uploadResource := newUploadResource()
	container.Add(uploadResource.WebService())

	projectsResource := newProjectsResource()
	container.Add(projectsResource.WebService())

	return container
}
