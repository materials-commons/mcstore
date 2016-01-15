package mccli

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var downloadDirCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Download a project directory",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "parallel, n",
			Value: 3,
			Usage: "Number of simultaneous downloads to perform, defaults to 3",
		},
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project to download the file from",
		},
	},
	Action: downloadDirCLI,
}

func downloadDirCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a directory to download.")
		os.Exit(1)
	}

	dirPath := filepath.Clean(c.Args()[0])
	if !validateDirectoryPath(dirPath) {
		os.Exit(1)
	}

	project := c.String("project")
	numThreads := getNumThreads(c)
	recursive := c.Bool("recursive")
	client := mc.NewClientAPI()

	if err := client.DownloadDirectory(project, dirPath, recursive, numThreads); err != nil {
		fmt.Println("Directory download failed:", err)
		os.Exit(1)
	}

	os.Exit(1)
}
