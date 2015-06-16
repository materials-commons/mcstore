package ws

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
)

// HTTPError is the error and message to respond with.
type HTTPError struct {
	statusCode int
	message    string
}

// Write writes the HTTPError to the ResponseWriter
func (e *HTTPError) Write(response http.ResponseWriter) {
	r, ok := response.(*restful.Response)
	if ok {
		r.WriteErrorString(e.statusCode, e.message)
	} else {
		response.WriteHeader(e.statusCode)
		response.Write([]byte(e.message))
	}
}

// WriteError writes the specified error to the writer. It translates the
// error to a HTTP status code.
func WriteError(err error, writer http.ResponseWriter) {
	httpErr := ErrorToHTTPError(err)
	httpErr.Write(writer)
}

// ErrorToHTTPError translates an error to an HTTPError. It translates an
// error to an HTTP status code.
func ErrorToHTTPError(err error) *HTTPError {
	switch e := err.(type) {
	case *app.Error:
		return appErrToHTTPError(e)
	default:
		return otherErrorToHTTPError(e)
	}
}

// appToHTTPError tranlates an mcerr.Error to an httpError.
func appErrToHTTPError(err *app.Error) *HTTPError {
	httpErr := otherErrorToHTTPError(err.Err)
	httpErr.message = fmt.Sprintf("%s: %s", httpErr.message, err.Message)
	return httpErr
}

// otherErrorToHTTPError translates other error types to an httpError.
func otherErrorToHTTPError(err error) *HTTPError {
	var httpErr HTTPError
	switch err {
	case app.ErrNotFound:
		httpErr.statusCode = http.StatusBadRequest
	case app.ErrInvalid:
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
