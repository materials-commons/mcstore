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
	switch {
	case len(errs) != 0:
		return app.ErrInvalid
	case resp.StatusCode == http.StatusInternalServerError:
		return app.ErrInternal
	case resp.StatusCode == http.StatusBadRequest:
		return app.ErrInvalid
	case resp.StatusCode == http.StatusNotFound:
		return app.ErrNotFound
	case resp.StatusCode == http.StatusForbidden:
		return app.ErrExists
	case resp.StatusCode == http.StatusUnauthorized:
		return app.ErrNoAccess
	case resp.StatusCode > 299:
		app.Log.Errorf("Unclassified error %d: %s", resp.StatusCode, resp.Status)
		return app.ErrUnclassified
	default:
		return nil
	}
}

func ToJSON(from string, to interface{}) error {
	err := json.Unmarshal([]byte(from), to)
	return err
}
