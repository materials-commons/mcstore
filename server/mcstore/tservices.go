package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/testdb"
)

// NewServicesContainerForTest creates a version of the container that
// connects to the test database.
func NewServicesContainerForTest() *restful.Container {
	container := restful.NewContainer()
	databaseSessionFilter := &databaseSessionFilter{
		session: testdb.RSession,
	}
	container.Filter(databaseSessionFilter.Filter)

	apikeyFilter := newAPIKeyFilter(apiKeyCache)
	container.Filter(apikeyFilter.Filter)

	// launch routine to track changes to users and
	// update the keycache appropriately.
	go updateKeyCacheOnChange(testdb.RSessionMust(), apiKeyCache)

	uploadResource := newUploadResource()
	container.Add(uploadResource.WebService())

	projectsResource := newProjectsResource()
	container.Add(projectsResource.WebService())
	return container
}
