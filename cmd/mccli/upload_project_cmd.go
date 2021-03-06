package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var uploadProjectCommand = cli.Command{
	Name:    "project",
	Aliases: []string{"proj", "p"},
	Usage:   "Upload project to MaterialsCommons",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "parallel, n",
			Value: 3,
			Usage: "Number of simultaneous uploads to perform, defaults to 3",
		},
	},
	Action: uploadProjectCLI,
}

// uploadProjectCLI implements the cli command upload project.
func uploadProjectCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project to upload.")
		os.Exit(1)
	}

	projectName := c.Args()[0]
	numThreads := getNumThreads(c)

	client := mc.NewClientAPI()
	if err := client.UploadProject(projectName, numThreads); err != nil {
		fmt.Println("Project upload failed:", err)
		os.Exit(1)
	}

	fmt.Println("Project successfully uploaded.")
}
