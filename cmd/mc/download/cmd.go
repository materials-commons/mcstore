package download

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func Cmd(c *cli.Context) {
	fmt.Println("download: ", c.Args())
}
