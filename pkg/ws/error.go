package ws

import (
	"fmt"
	"net/http"

	"github.com/materials-commons/mcstore/pkg/app"
)

// HTTPError is the error and message to respond with.
type HTTPError struct {
	statusCode int
	message    string
}

func (e *HTTPError) Write(response http.ResponseWriter) {
	response.WriteHeader(e.statusCode)
	response.Write([]byte(e.message))
}

// if the error code is of type mcerr.Error and handles it appropriately.
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
