package mc

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var ReceiveCommand = cli.Command{
	Name:    "receive",
	Aliases: []string{"rec", "r"},
	Usage:   "Receive data over the air",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
	},
	Action: receiveCLI,
}

func receiveCLI(c *cli.Context) {
	fmt.Println("receive: ", c.Args())
}
