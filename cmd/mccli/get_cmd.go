package mccli

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var GetCommand = cli.Command{
	Name:  "get",
	Usage: "Get a property",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
	},
	Action: getCLI,
}

func getCLI(c *cli.Context) {
	fmt.Println("get: ", c.Args())
}
