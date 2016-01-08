package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var downloadProjectCommand = cli.Command{
	Name:    "project",
	Aliases: []string{"proj", "p"},
	Usage:   "Download all files and directories in a project",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "parallel, n",
			Value: 3,
			Usage: "Number of simultaneous downloads to perform, defaults to 3",
		},
	},
	Action: downloadProjectCLI,
}

func downloadProjectCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project to download.")
		os.Exit(1)
	}

	projectName := c.Args()[0]
	numThreads := getNumThreads(c)
	client := mc.NewClientAPI()

	if err := client.DownloadProject(projectName, numThreads); err != nil {
		fmt.Println("Project download failed:", err)
		os.Exit(1)
	}

	fmt.Println("Project successfully downloaded.")
}
