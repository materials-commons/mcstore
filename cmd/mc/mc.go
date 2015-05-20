package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/config"
	"github.com/materials-commons/config/cfg"
	"github.com/materials-commons/config/handler"
	"github.com/materials-commons/config/loader"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/cmd/mc/download"
	"github.com/materials-commons/mcstore/cmd/mc/get"
	"github.com/materials-commons/mcstore/cmd/mc/login"
	"github.com/materials-commons/mcstore/cmd/mc/monitor"
	"github.com/materials-commons/mcstore/cmd/mc/receive"
	"github.com/materials-commons/mcstore/cmd/mc/send"
	"github.com/materials-commons/mcstore/cmd/mc/set"
	"github.com/materials-commons/mcstore/cmd/mc/upload"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "V. Glenn Tarcea",
			Email: "gtarcea@umich.edu",
		},
	}
	app.Commands = []cli.Command{
		send.Command,
		receive.Command,
		get.Command,
		set.Command,
		upload.Command,
		download.Command,
		monitor.Command,
		login.Command,
	}

	setupConfig()
	app.Run(os.Args)
}

// setupConfig sets up config for the process. It creates a cascade
// of config handlers to search.
func setupConfig() {
	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Couldn't determine current user: %s", err))
	}

	handlers := []cfg.Handler{
		handler.Env(),
	}

	configFile := filepath.Join(u.HomeDir, ".materialscommons/config.json")
	loader := getConfigLoader(configFile)
	if loader != nil {
		handlers = append(handlers, handler.Loader(loader))
	}

	defaultHandler := handler.Map()
	loadDefaults(defaultHandler)
	handlers = append(handlers, defaultHandler)

	config.Init(handler.Sync(handler.Multi(handlers...)))
}

// getConfigLoader returns a json loader if the $HOME/.materialscommons/config.json
// file exists. It will panic if the file exists but cannot be read.
func getConfigLoader(configFile string) cfg.Loader {
	if file.Exists(configFile) {
		contents, err := ioutil.ReadFile(configFile)
		if err != nil {
			panic(fmt.Sprintf("%s exists but can't be read: %s", configFile, err))
		}
		return loader.JSON(bytes.NewReader(contents))
	}
	return nil
}

// loadDefaults sets up the default values for the following configuration keys:
//     mcurl: https://materialscommons.org/api
func loadDefaults(h cfg.Handler) {
	h.Set("mcurl", "https://materialscommons.org/api")
}
