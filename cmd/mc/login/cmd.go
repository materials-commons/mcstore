package login

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var Command = cli.Command{
	Name:    "login",
	Aliases: []string{"l"},
	Usage:   "Login to MaterialsCommons",
	Flags:   []cli.Flag{},
	Action:  Cmd,
}

func Cmd(c *cli.Context) {
	fmt.Println("login: ", c.Args())
}
