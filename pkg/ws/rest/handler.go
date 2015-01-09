package rest

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
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
type RouteFunc func(request *restful.Request, response *restful.Response, user schema.User) (error, interface{})

// RouteFunc1 is a route function that only returns an error, but no value
type RouteFunc1 func(request *restful.Request, response *restful.Response, user schema.User) error

// Handler represents the way a route function should actually be written.
type Handler func(request *restful.Request, response *restful.Response)

// RouteHandler creates a wrapper function for route methods. This allows route
// methods to return errors and have them handled correctly.
func RouteHandler(f RouteFunc) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		user := schema.User{} //request.Attribute("user").(schema.User)
		err, val := f(request, response, user)
		switch {
		case err != nil:
			httpErr := errorToHTTPError(err)
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
	f2 := func(request *restful.Request, response *restful.Response, user schema.User) (error, interface{}) {
		err := f(request, response, user)
		return err, nil
	}
	return RouteHandler(f2)
}

// ErrorToHTTPError translates an error code into an httpError. It checks
// if the error code is of type mcerr.Error and handles it appropriately.
func ErrorToHTTPError(err error) *httpError {
	switch e := err.(type) {
	case *app.Error:
		return appErrToHTTPError(e)
	default:
		return otherErrorToHTTPError(e)
	}
}

// appErrToHTTPError tranlates an app.Error to an httpError.
func appErrToHTTPError(err *app.Error) *httpError {
	httpErr := otherErrorToHTTPError(err.Err)
	httpErr.message = fmt.Sprintf("%s: %s", httpErr.message, err.Message)
	return httpErr
}

// otherErrorToHTTPError translates other error types to an httpError.
func otherErrorToHTTPError(err error) *httpError {
	var httpErr httpError
	switch err {
	case app.ErrNotFound:
		httpErr.statusCode = http.StatusBadRequest
	case app.ErrExists:
		httpErr.statusCode = http.StatusForbidden
	case app.ErrNoAccess:
		httpErr.statusCode = http.StatusUnauthorized
	default:
		httpErr.statusCode = http.StatusInternalServerError
	}

	httpErr.message = err.Error()
	return &httpErr
}
