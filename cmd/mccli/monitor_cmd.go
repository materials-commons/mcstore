package mccli

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var MonitorCommand = cli.Command{
	Name:    "monitor",
	Aliases: []string{"mon", "m"},
	Usage:   "Monitor a directory for changes",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
	},
	Action: monitorCLI,
}

func monitorCLI(c *cli.Context) {
	fmt.Println("monitor: ", c.Args())
}
