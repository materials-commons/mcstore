package send

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func Cmd(c *cli.Context) {
	fmt.Println("send: ", c.Args())
}
