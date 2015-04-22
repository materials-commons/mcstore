package login

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func Cmd(c *cli.Context) {
	fmt.Println("login: ", c.Args())
}
