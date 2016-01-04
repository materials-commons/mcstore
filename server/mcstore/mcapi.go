package mcstore

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/mcstore/pkg/app"
	"gnd.la/net/urlutil"
)

// MCUrl returns the current mcurl config entry.
func MCUrl() string {
	return config.GetString("mcurl")
}

// MCClient creates a new EzClient.
func MCClient() *ezhttp.EzClient {
	mcurl := MCUrl()
	if strings.HasPrefix(mcurl, "https") {
		return ezhttp.NewSSLClient()
	}
	return ezhttp.NewClient()
}

// Url create the url for accessing a service. It adds the mcurl to
// the path, and also adds the apikey argument.
func Url(path string) string {
	values := url.Values{}
	values.Add("apikey", config.GetString("apikey"))
	mcurl := MCUrl()
	if strings.HasPrefix(path, "/") {
		mcurl = MCUrl() + path
	} else {
		mcurl = MCUrl() + "/" + path
	}
	mcurl = urlutil.AppendQuery(mcurl, values)
	return mcurl
}

// ToError tests the list of errors and the response to determine
// the type of error to return. It calls HTTPStatusToError to
// translate response status codes to an error.
func ToError(resp *http.Response, errs []error) error {
	if len(errs) != 0 {
		return app.ErrInvalid
	}
	return HTTPStatusToError(resp.StatusCode)
}

// HTTPStatusToError translates an http state to an
// application error.
func HTTPStatusToError(status int) error {
	switch {
	case status == http.StatusInternalServerError:
		return app.ErrInternal
	case status == http.StatusBadRequest:
		return app.ErrInvalid
	case status == http.StatusNotFound:
		return app.ErrNotFound
	case status == http.StatusForbidden:
		return app.ErrExists
	case status == http.StatusUnauthorized:
		return app.ErrNoAccess
	case status > 299:
		app.Log.Errorf("Unclassified error %d", status)
		return app.ErrUnclassified
	default:
		return nil
	}
}

// ToJSON unmarshalls a string that contains JSON.
func ToJSON(from string, to interface{}) error {
	err := json.Unmarshal([]byte(from), to)
	return err
}
