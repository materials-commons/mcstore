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

func MCUrl() string {
	return config.GetString("mcurl")
}

func MCClient() *ezhttp.EzClient {
	mcurl := MCUrl()
	if strings.HasPrefix(mcurl, "https") {
		return ezhttp.NewSSLClient()
	}
	return ezhttp.NewClient()
}

func Url(path string) string {
	values := url.Values{}
	values.Add("apikey", config.GetString("apikey"))
	mcurl := urlutil.MustJoin(MCUrl(), path)
	mcurl = urlutil.AppendQuery(mcurl, values)
	return mcurl
}

func ToError(resp *http.Response, errs []error) error {
	if len(errs) != 0 {
		return app.ErrInvalid
	}
	return HTTPStatusToError(resp.StatusCode)
}

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

func ToJSON(from string, to interface{}) error {
	err := json.Unmarshal([]byte(from), to)
	return err
}
