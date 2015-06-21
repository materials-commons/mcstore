package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/testdb"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/uploads"
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

	uploadResource := createUploadsResourceForTest()
	container.Add(uploadResource.WebService())

	projectsResource := createProjectsResourceForTest()
	container.Add(projectsResource.WebService())
	return container
}

func createUploadsResourceForTest() rest.Service {
	session := testdb.RSession()
	return newUploadResource(uploads.NewUploadServiceUsingSession(session),
		uploads.NewIDServiceUsingSession(session), newDirServiceUsingSession(session))
}

func createProjectsResourceForTest() rest.Service {
	session := testdb.RSession()
	return newProjectsResource(newDirServiceUsingSession(session), newProjectServiceUsingSession(session))
}
