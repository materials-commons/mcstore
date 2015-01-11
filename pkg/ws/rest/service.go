package rest

import "github.com/emicklei/go-restful"

// Service implements a REST based web service.
type Service interface {
	WebService() *restful.WebService
}
