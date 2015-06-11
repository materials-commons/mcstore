package mc

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var LoginCommand = cli.Command{
	Name:    "login",
	Aliases: []string{"l"},
	Usage:   "Login to MaterialsCommons",
	Flags:   []cli.Flag{},
	Action:  loginCLI,
}

func loginCLI(c *cli.Context) {
	fmt.Println("login: ", c.Args())
}
