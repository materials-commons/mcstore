package client

import (
	"crypto/tls"

	"github.com/parnurzeal/gorequest"
)

func NewGoRequest() *gorequest.SuperAgent {
	return gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}
