package mccli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/server/mcstore"
)

var ShowCommand = cli.Command{
	Name:   "show",
	Usage:  "Show the configuration",
	Action: showCLI,
}

func showCLI(c *cli.Context) {
	apikey := config.GetString("apikey")
	mcurl := mcstore.MCUrl()
	mclogging := config.GetString("mclogging")
	fmt.Println("apikey:", apikey)
	fmt.Println("mcurl:", mcurl)
	fmt.Println("mclogging:", mclogging)
}
