package mccli

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var DownloadCommand = cli.Command{
	Name:    "download",
	Aliases: []string{"down", "d"},
	Usage:   "Download data from MaterialsCommons",
	Flags:   []cli.Flag{},
	Action:  downloadCLI,
}

func downloadCLI(c *cli.Context) {
	fmt.Println("download: ", c.Args())
}
