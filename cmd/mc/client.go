package mc

import (
	"crypto/tls"

	"github.com/parnurzeal/gorequest"
)

func newGoRequest() *gorequest.SuperAgent {
	return gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}
