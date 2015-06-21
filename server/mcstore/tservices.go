package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/testdb"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
)

// NewServicesContainerForTest creates a version of the container that
// connects to the test database.
func NewServicesContainerForTest() *restful.Container {
	container := restful.NewContainer()
	databaseSessionFilter := &databaseSessionFilter{
		session: testdb.RSessionErr,
	}
	container.Filter(databaseSessionFilter.Filter)

	apikeyFilter := newAPIKeyFilter()
	container.Filter(apikeyFilter.Filter)

	uploadResource := newUploadResource()
	container.Add(uploadResource.WebService())

	projectsResource := createProjectsResourceForTest()
	container.Add(projectsResource.WebService())
	return container
}

func createProjectsResourceForTest() rest.Service {
	session := testdb.RSession()
	return newProjectsResource(newDirServiceUsingSession(session), newProjectServiceUsingSession(session))
}
