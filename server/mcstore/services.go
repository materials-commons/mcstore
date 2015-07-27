package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db"
)

// NewServicesContainer creates a new restful.Container made up of all
// the rest resources handled by the server.
func NewServicesContainer(sc db.SessionCreater) *restful.Container {
	container := restful.NewContainer()

	databaseSessionFilter := &databaseSessionFilter{
		session: sc.RSession,
	}
	container.Filter(databaseSessionFilter.Filter)

	apikeyFilter := newAPIKeyFilter(apiKeyCache)
	container.Filter(apikeyFilter.Filter)

	// launch routine to track changes to users and
	// update the keycache appropriately.
	go updateKeyCacheOnChange(sc.RSessionMust(), apiKeyCache)

	uploadResource := newUploadResource()
	container.Add(uploadResource.WebService())

	projectsResource := newProjectsResource()
	container.Add(projectsResource.WebService())

	searchResource := newSearchResource()
	container.Add(searchResource.WebService())

	return container
}
