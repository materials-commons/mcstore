package mc

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var SetCommand = cli.Command{
	Name:  "set",
	Usage: "Set property",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
	},
	Action: setCLI,
}

func setCLI(c *cli.Context) {
	fmt.Println("set: ", c.Args())
}
