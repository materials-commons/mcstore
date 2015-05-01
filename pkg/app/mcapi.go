package app

import (
	"strings"

	"github.com/materials-commons/config"
	"github.com/materials-commons/gohandy/ezhttp"
)

type mcapi struct{}

var MCApi mcapi

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
