package get

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:  "get",
	Usage: "Get a property",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
	},
	Action: Cmd,
}

func Cmd(c *cli.Context) {
	fmt.Println("get: ", c.Args())
}
