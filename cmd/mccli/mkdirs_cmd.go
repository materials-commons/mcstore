package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var (
	MkdirsCommand = cli.Command{
		Name:    "mkdirs",
		Aliases: []string{"mkd"},
		Usage:   "Create project directories",
		Action:  mkdirsCLI,
	}
)

func mkdirsCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project name")
		os.Exit(1)
	}

	projectName := c.Args()[0]

	client := mc.NewClientAPI()
	if err := client.CreateProjectDirectories(projectName); err != nil {
		fmt.Println("Unable to create project directories:", err)
		os.Exit(1)
	}

	fmt.Println("Project directories created")
}
