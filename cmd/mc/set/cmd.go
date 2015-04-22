package set

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:  "set",
	Usage: "Set property",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
	},
	Action: Cmd,
}

func Cmd(c *cli.Context) {
	fmt.Println("set: ", c.Args())
}
