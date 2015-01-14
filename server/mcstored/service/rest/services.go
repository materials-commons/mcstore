package rest

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstored/service/rest/upload"
)

func NewServicesContainer() *restful.Container {
	container := restful.NewContainer()
	uploadResource := uploadResource()
	container.Add(uploadResource.WebService())
	return container
}

func uploadResource() rest.Service {
	tracker := upload.NewUploadTracker()
	finisherFactory := upload.NewUploadFinisherFactory(tracker)
	assemblerFactory := upload.NewMCDirAssemblerFactory(finisherFactory)
	rw := upload.NewFileRequestWriter(upload.NewMCDirRequestPath())
	uploader := upload.NewUploader(rw, tracker)
	return upload.NewResource(uploader, assemblerFactory)
}
