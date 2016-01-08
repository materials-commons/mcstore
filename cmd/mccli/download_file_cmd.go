package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

var downloadFileCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Download a project file",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project to download the file from",
		},
	},
	Action: downloadFileCLI,
}

func downloadFileCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a file to download.")
		os.Exit(1)
	}
}
