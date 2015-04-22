package send

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:    "send",
	Aliases: []string{"s"},
	Usage:   "Send data over the air",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "name of receiving service",
		},
		cli.StringFlag{
			Name:  "project, p",
			Usage: "project to send from",
		},
		cli.StringFlag{
			Name:  "directory, d",
			Usage: "directory to send files from",
		},
		cli.StringFlag{
			Name:  "file, f",
			Usage: "file to send",
		},
	},
	Action: Cmd,
}

func Cmd(c *cli.Context) {
	fmt.Println("send: ", c.Args())
}
