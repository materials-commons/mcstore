package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

var downloadDirCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Download a project directory",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "parallel, n",
			Value: 3,
			Usage: "Number of simultaneous downloads to perform, defaults to 3",
		},
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project to download the file from",
		},
	},
	Action: downloadDirCLI,
}

func downloadDirCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a directory to download.")
		os.Exit(1)
	}
}
