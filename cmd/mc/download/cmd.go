package download

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:    "download",
	Aliases: []string{"down", "d"},
	Usage:   "Download data from MaterialsCommons",
	Flags:   []cli.Flag{},
	Action:  Cmd,
}

func Cmd(c *cli.Context) {
	fmt.Println("download: ", c.Args())
}
