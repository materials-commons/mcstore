package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/testdb"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/uploads"
)

// NewServicesContainer creates a new restful.Container made up of all
// the rest resources handled by the server.
func NewServicesContainer() *restful.Container {
	container := restful.NewContainer()

	apikeyFilter := newAPIKeyFilter(dai.NewRUsers(db.RSessionMust()))
	container.Filter(apikeyFilter.Filter)
	// to add in filter for database sessions the code would look like
	// the following. Note that this assumes we have changed the rest of
	// the code to get the session from the variables.
	// dbSessionFilter := &databaseSessionFilter
	// container.Filter(dbSessionFilter.Filter)
	// container.Filter(apikeyFilter.Filter)

	uploadResource := createUploadsResource()
	container.Add(uploadResource.WebService())

	projectsResource := createProjectsResource()
	container.Add(projectsResource.WebService())

	return container
}

// uploadResource creates a new upload resource.
func createUploadsResource() rest.Service {
	return newUploadResource(uploads.NewUploadService(), uploads.NewIDService(), newDirService())
}

// projectsResource creates a new projects resource.
func createProjectsResource() rest.Service {
	return newProjectsResource(newDirService(), newProjectService())
}

func NewServicesContainerForTest() *restful.Container {
	container := restful.NewContainer()
	apikeyFilter := newAPIKeyFilter(dai.NewRUsers(testdb.RSession()))
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
