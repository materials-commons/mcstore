package mc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/mcstore/pkg/app"
	"gnd.la/net/urlutil"
)

type mcapi struct{}

var Api mcapi

func (a mcapi) MCUrl() string {
	return config.GetString("mcurl")
}

func (a mcapi) MCClient() *ezhttp.EzClient {
	mcurl := a.MCUrl()
	if strings.HasPrefix(mcurl, "https") {
		return ezhttp.NewSSLClient()
	}
	return ezhttp.NewClient()
}

func (a mcapi) Url(path string) string {
	values := url.Values{}
	values.Add("apikey", config.GetString("apikey"))
	mcurl := urlutil.MustJoin(a.MCUrl(), path)
	mcurl = urlutil.AppendQuery(mcurl, values)
	return mcurl
}

func (a mcapi) IsError(resp *http.Response, errs []error) error {
	switch {
	case len(errs) != 0:
		return app.ErrInvalid
	case resp.StatusCode > 299:
		return fmt.Errorf("HTTP Error: %s", resp.Status)
	default:
		return nil
	}
}

func (a mcapi) ToJSON(from string, to interface{}) error {
	err := json.Unmarshal([]byte(from), to)
	return err
}
