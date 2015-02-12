package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/server/options"
)

type opts struct {
	options.Standard
}

// init initializes config for the server.
func init() {
	config.Init(config.TwelveFactorWithOverride)
	config.SetErrorHandler(configErrorHandler)
}

// configErrorHandler gives us a chance to handle configuration look up errors.
func configErrorHandler(key string, err error, args ...interface{}) {

}

func main() {
	var opts opts
	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	opts.Perform()
	server(opts.Server.HTTPPort)
}

func server(port uint) {

}
