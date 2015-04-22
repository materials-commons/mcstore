package upload

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func Cmd(c *cli.Context) {
	fmt.Println("upload: ", c.Args())
}
