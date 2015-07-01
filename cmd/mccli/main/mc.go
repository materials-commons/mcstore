package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/config"
	"github.com/materials-commons/config/cfg"
	"github.com/materials-commons/config/handler"
	"github.com/materials-commons/config/loader"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/cmd/mccli"
	"github.com/materials-commons/mcstore/pkg/app"
)

// init sets up the package by loading and configuring the config package.
func init() {
	handler := setupConfigHandler()
	config.Init(handler)
	setLoggingLevel()
}

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
		mccli.CreateCommand,
		mccli.SetupCommand,
		mccli.ShowCommand,
		mccli.UploadCommand,
	}
	app.Run(os.Args)
}

// setupConfigHandler creates the handler for the mc package. It sets up a
// multi handler. If the user has setup a config.json in their .materialscommons
// directory then it will add that to the handler list.
func setupConfigHandler() cfg.Handler {
	handlers := []cfg.Handler{
		handler.Env(),
	}

	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Couldn't determine current user: %s", err))
	}

	handlers := []cfg.Handler{
		handler.Env(),
	}

	configFile := filepath.Join(u.HomeDir, ".materialscommons/config.json")
	loader := getUserConfigLoader(configFile)
	if loader != nil {
		handlers = append(handlers, handler.Loader(loader))
	}

	defaultHandler := handler.Map()
	loadDefaults(defaultHandler)
	handlers = append(handlers, defaultHandler)
	return handler.Sync(handler.Multi(handlers...))
}

// getUserConfigLoader returns a json loader if the $HOME/.materialscommons/config.json
// file exists. It will panic if the file exists but cannot be read.
func getUserConfigLoader(configFile string) cfg.Loader {
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
//     mclogging: info
func loadDefaults(h cfg.Handler) {
	h.Set("mcurl", "https://materialscommons.org/api")
	h.Set("mclogging", "info")
}

// setLoggingLevel sets the apps logging level.
func setLoggingLevel() {
	level := config.GetString("mclogging")
	if lvl, err := log15.LvlFromString(level); err != nil {
		fmt.Printf("Invalid Log Level: %s, setting to info.\n", level)
		fmt.Println("Valid logging levels are: debug, info, warn, error, crit. The default is info.")
		app.SetLogLvl(log15.LvlInfo)
	} else {
		app.SetLogLvl(lvl)
	}
}
