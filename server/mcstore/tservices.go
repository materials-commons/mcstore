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
		session: testdb.RSessionErr,
	}
	container.Filter(databaseSessionFilter.Filter)

	apikeyFilter := newAPIKeyFilter()
	container.Filter(apikeyFilter.Filter)

	uploadResource := newUploadResource()
	container.Add(uploadResource.WebService())

	projectsResource := newProjectsResource()
	container.Add(projectsResource.WebService())
	return container
}
