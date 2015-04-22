package monitor

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:    "monitor",
	Aliases: []string{"mon", "m"},
	Usage:   "Monitor a directory for changes",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
	},
	Action: Cmd,
}

func Cmd(c *cli.Context) {
	fmt.Println("monitor: ", c.Args())
}
