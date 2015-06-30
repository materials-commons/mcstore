package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var (
	projectCreateCommand = cli.Command{
		Name:    "create",
		Aliases: []string{"cr", "c"},
		Usage:   "Create a new project",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "directory, dir, d",
				Usage: "The base directory for the project",
			},
		},
		Action: projectCreateCLI,
	}
)

// projectCreateCLI implements the project create command.
func projectCreateCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project name")
		os.Exit(1)
	}
	projectName := c.Args()[0]

	dirPath := c.String("directory")
	if !validateDirectoryPath(dirPath) {
		os.Exit(1)
	}

	client := mc.NewClientAPI()
	if err := client.CreateProject(projectName, dirPath); err != nil {
		fmt.Println("Unable to create project:", err)
		os.Exit(1)
	}

	fmt.Println("Project successfully created")
}

// validateDirectoryPath checks that the given directory path exists.
func validateDirectoryPath(path string) bool {
	if path == "" {
		fmt.Println("You must specify a local directory path where the project files are located.")
		return false
	}

	if _, err := os.Stat(path); err != nil {
		fmt.Println("Directory doesn't exist or you don't have access")
		return false
	}

	return true
}
