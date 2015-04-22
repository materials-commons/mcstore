package upload

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:    "upload",
	Aliases: []string{"up", "u"},
	Usage:   "Upload data to MaterialsCommons",
	Flags:   []cli.Flag{},
	Action:  Cmd,
}

func Cmd(c *cli.Context) {
	fmt.Println("upload: ", c.Args())
}
