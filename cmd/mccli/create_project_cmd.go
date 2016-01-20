package mccli

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var (
	createProjectCommand = cli.Command{
		Name:    "project",
		Aliases: []string{"proj", "p"},
		Usage:   "Create a new project",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "directory, dir, d",
				Usage: "The base directory for the project",
			},
			cli.BoolFlag{
				Name:  "upload, up, u",
				Usage: "Upload project after creating it",
			},
			cli.IntFlag{
				Name:  "parallel, n",
				Value: 3,
				Usage: "Number of simultaneous uploads to perform, defaults to 3",
			},
			cli.BoolFlag{
				Name:  "download, down, dl",
				Usage: "Download the projects files after creating it",
			},
		},
		Action: createProjectCLI,
	}
)

// createProjectCLI implements the create project command.
func createProjectCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project name")
		os.Exit(1)
	}
	projectName := c.Args()[0]

	dirPath := c.String("directory")
	dirPath, err := filepath.Abs(dirPath)
	if err != nil {
		fmt.Println("Unable to create absolute directory path: ", err)
		os.Exit(1)
	}

	dirPath = filepath.Clean(dirPath)

	if !validateDirectoryPath(dirPath) {
		os.Exit(1)
	}

	client := mc.NewClientAPI()
	if err := client.CreateProject(projectName, dirPath); err != nil {
		fmt.Println("Unable to create project:", err)
		os.Exit(1)
	}

	fmt.Println("Project successfully created.")
	numThreads := getNumThreads(c)

	if c.Bool("upload") {
		if err := client.UploadProject(projectName, numThreads); err != nil {
			fmt.Println("Project upload failed:", err)
			os.Exit(1)
		}
		fmt.Println("Project successfully uploaded.")
	}

	if c.Bool("download") {
		if err := client.DownloadProject(projectName, numThreads); err != nil {
			fmt.Println("Project download failed:", err)
			os.Exit(1)
		}
		fmt.Println("Project successfully downloaded.")
	}
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
