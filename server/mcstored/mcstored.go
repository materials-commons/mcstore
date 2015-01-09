// Package mcstored implements the server for storage requests.
package mcstored

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
)

// Options for server startup
type serverOptions struct {
	MCDir    string `long:"mcdir" description:"Directory path to materials commons file storage"`
	PrintPid bool   `long:"print-pid" description:"Prints the server pid to stdout"`
	HTTPPort uint   `long:"http-port" description:"Port webserver listens on" default:"5010"`
}

// Options for the database
type databaseOptions struct {
	Connection string `long:"db-connect" description:"The database connection string"`
	Name       string `long:"db" description:"Database to use"`
	Type       string `long:"db-type" description:"The type of database to connect to"`
}

// Break the options into option groups.
type options struct {
	Server   serverOptions   `group:"Server Options"`
	Database databaseOptions `group:"Database Options"`
}

func configErrorHandler(key string, err error, args ...interface{}) {

}

func init() {
	config.Init(config.TwelveFactorWithOverride)
	config.SetErrorHandler(configErrorHandler)
}

func main() {
	var opts options
	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	if opts.Server.PrintPid {
		fmt.Println(os.Getpid())
	}

	setupConfig(opts)
	server(opts.Server.HTTPPort)
}

func setupConfig(opts options) {
	if opts.Database.Connection != "" {
		config.Set("MCDB_CONNECTION", opts.Database.Connection)
	}

	if opts.Database.Name != "" {
		config.Set("MCDB_NAME", opts.Database.Name)
	}

	if opts.Database.Type != "" {
		config.Set("MCDB_TYPE", opts.Database.Type)
	}

	if opts.Server.MCDir != "" {
		config.Set("MCDIR", opts.Server.MCDir)
	}
}

func server(port uint) {
	http.HandleFunc("/data/", dataHandler)
	app.Log.Crit("http Server failed", "error", http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func dataHandler(writer http.ResponseWriter, req *http.Request) {

}
