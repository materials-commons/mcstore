package rest

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstored/service/data"
	"github.com/materials-commons/mcstore/server/mcstored/service/rest/filters"
	"github.com/materials-commons/mcstore/server/mcstored/service/rest/upload"
	"github.com/materials-commons/mcstore/server/mcstored/service/uploads"
)

// NewServicesContainer creates a new restful.Container made up of all
// the rest resources handled by the server.
func NewServicesContainer() *restful.Container {
	apikeyFilter := filters.NewAPIKeyFilter(dai.NewRUsers(db.RSessionMust()))
	container := restful.NewContainer()
	container.Filter(apikeyFilter.Filter)
	uploadResource := uploadResource()
	container.Add(uploadResource.WebService())
	return container
}

// uploadResource creates a new upload resource.
func uploadResource() rest.Service {
	return upload.NewResource(uploads.NewUploadService(), uploads.NewIDService(), data.NewDirService())
}
