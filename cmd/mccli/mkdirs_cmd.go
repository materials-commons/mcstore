package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var (
	mkdirsProjectCommand = cli.Command{
		Name:    "project",
		Aliases: []string{"proj", "p"},
		Usage:   "Create project directories on MaterialsCommons",
		Action:  mkdirsProjectCLI,
	}

	MkdirsCommand = cli.Command{
		Name:    "mkdirs",
		Aliases: []string{"mkd"},
		Usage:   "Create project directories",
		Subcommands: []cli.Command{
			mkdirsProjectCommand,
		},
	}
)

func mkdirsProjectCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project name")
		os.Exit(1)
	}

	projectName := c.Args()[0]

	client := mc.NewClientAPI()
	if err := client.CreateProjectDirectories(projectName); err != nil {
		fmt.Printf("Unable to create project directories for project %s: %s\n", projectName, err)
		os.Exit(1)
	}

	fmt.Println("Project directories created")
}
