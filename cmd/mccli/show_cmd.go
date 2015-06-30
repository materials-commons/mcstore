package mccli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/server/mcstore"
)

var ShowCommand = cli.Command{
	Name:    "show",
	Aliases: []string{"sh"},
	Usage:   "Show commands",
	Subcommands: []cli.Command{
		showConfigCommand,
	},
}

var showConfigCommand = cli.Command{
	Name:    "config",
	Aliases: []string{"conf", "c"},
	Usage:   "Show configuration",
	Action:  showConfigCLI,
}

func showConfigCLI(c *cli.Context) {
	apikey := config.GetString("apikey")
	mcurl := mcstore.MCUrl()
	mclogging := config.GetString("mclogging")
	fmt.Println("apikey:", apikey)
	fmt.Println("mcurl:", mcurl)
	fmt.Println("mclogging:", mclogging)
}
