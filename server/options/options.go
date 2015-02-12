package options

import (
	"fmt"
	"os"

	"github.com/inconshreveable/log15"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
)

// Options for the database
type Database struct {
	Connection string `long:"db-connect" description:"The database connection string"`
	Name       string `long:"db" description:"Database to use" default:"materialscommons"`
}

// Options for server startup
type Server struct {
	ShowPid  bool   `long:"show-pid" description:"Prints the server pid to stdout"`
	HTTPPort uint   `long:"http-port" description:"Port webserver listens on" default:"5010"`
	LogLevel string `long:"log-level" description:"Logging level for server (debug, info, warn, error, crit)" default:"info"`
}

type Standard struct {
	Database Database
	Server   Server
}

func (s Standard) Perform() {
	s.Database.setConfig()
	s.Server.setConfig()
	if s.Server.ShowPid {
		fmt.Println(os.Getpid())
	}
}

func (d Database) setConfig() {
	if d.Connection != "" {
		config.Set("MCDB_CONNECTION", d.Connection)
	}

	if d.Name != "" {
		config.Set("MCDB_NAME", d.Name)
	}
}

func (s Server) setConfig() {
	if lvl, err := log15.LvlFromString(s.LogLevel); err != nil {
		fmt.Printf("Invalid Log Level: %s, setting to Info\n", s.LogLevel)
		app.SetLogLvl(log15.LvlInfo)
	} else {
		fmt.Println("Log level set to:", s.LogLevel)
		app.SetLogLvl(lvl)
	}
}
