package mccli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
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
	fmt.Println("apikey:", config.GetString("apikey"))
	fmt.Println("mcurl:", mcstoreapi.MCUrl())
	fmt.Println("mclogging:", config.GetString("mclogging"))
}
