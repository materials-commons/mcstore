package rest

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws"
)

// httpError is the error and message to respond with.
type httpError struct {
	statusCode int
	message    string
}

// Write writes an httpError as the response.
func (e *httpError) Write(response *restful.Response) {
	response.WriteErrorString(e.statusCode, e.message)
}

// RouteFunc represents the routes function
type RouteFunc func(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error)

// RouteFunc1 is a route function that only returns an error, but no value
type RouteFunc1 func(request *restful.Request, response *restful.Response, user schema.User) error

// Handler represents the way a route function should actually be written.
type Handler func(request *restful.Request, response *restful.Response)

// RouteHandler creates a wrapper function for route methods. This allows route
// methods to return errors and have them handled correctly.
func RouteHandler(f RouteFunc) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		user := schema.User{} //request.Attribute("user").(schema.User)
		val, err := f(request, response, user)
		switch {
		case err != nil:
			httpErr := ws.ErrorToHTTPError(err)
			httpErr.Write(response)
		case val != nil:
			err = response.WriteEntity(val)
			if err != nil {
				// log the error here
			}
		default:
			// Nothing to do
		}
	}
}

// RouteHandler1 creates a wrapper function for route methods that only return an error. See
// RouteHandler for details.
func RouteHandler1(f RouteFunc1) restful.RouteFunction {
	// Create a function that looks like a RouteFunc but always returns null for its second return value
	f2 := func(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
		err := f(request, response, user)
		return nil, err
	}
	return RouteHandler(f2)
}
