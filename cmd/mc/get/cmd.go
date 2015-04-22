package get

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func Cmd(c *cli.Context) {
	fmt.Println("get: ", c.Args())
}
