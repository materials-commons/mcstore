// Package mcstored implements the server for storage requests.
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/inconshreveable/log15"
	"github.com/jessevdk/go-flags"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/materials-commons/mcstore/server/mcstored/service/rest"
	"github.com/materials-commons/mcstore/server/mcstored/service/web/content"
)

// Options for server startup
type serverOptions struct {
	MCDir    string `long:"mcdir" description:"Directory path to materials commons file storage"`
	PrintPid bool   `long:"print-pid" description:"Prints the server pid to stdout"`
	HTTPPort uint   `long:"http-port" description:"Port webserver listens on" default:"5010"`
	LogLevel string `long:"log-level" description:"Logging level for server (debug, info, warn, error, crit)" default:"info"`
}

// Options for the database
type databaseOptions struct {
	Connection string `long:"db-connect" description:"The database connection string"`
	Name       string `long:"db" description:"Database to use" default:"materialscommons"`
}

// Break the options into option groups.
type options struct {
	Server   serverOptions   `group:"Server Options"`
	Database databaseOptions `group:"Database Options"`
}

// configErrorHandler gives us a chance to handle configuration look up errors.
func configErrorHandler(key string, err error, args ...interface{}) {

}

// init initializes config for the server.
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

// setupConfig sets up configuration overrides that were passed in on the command line.
func setupConfig(opts options) {
	if opts.Database.Connection != "" {
		config.Set("MCDB_CONNECTION", opts.Database.Connection)
	}

	if opts.Database.Name != "" {
		config.Set("MCDB_NAME", opts.Database.Name)
	}

	if opts.Server.MCDir != "" {
		config.Set("MCDIR", opts.Server.MCDir)
	}

	if lvl, err := log15.LvlFromString(opts.Server.LogLevel); err != nil {
		fmt.Printf("Invalid Log Level: %s, setting to Info\n", opts.Server.LogLevel)
		app.SetLogLvl(log15.LvlInfo)
	} else {
		fmt.Println("Log level set to:", opts.Server.LogLevel)
		app.SetLogLvl(lvl)
	}
}

// server implements the actual serve for mcstored. It sets up the http routes and handlers. This
// method never returns.
func server(port uint) {
	container := rest.NewServicesContainer()
	http.Handle("/", container)

	session := db.RSessionMust()
	access := domain.NewAccess(dai.NewRProjects(session), dai.NewRFiles(session), dai.NewRUsers(session))
	dataHandler := content.NewDataHandler(access)
	http.Handle("/datafiles/static/", dataHandler)

	app.Log.Crit("http Server failed", "error", http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
