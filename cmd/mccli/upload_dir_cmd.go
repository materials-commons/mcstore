package mccli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var uploadDirCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Upload a file to MaterialsCommons",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project the file is in",
		},
		cli.BoolFlag{
			Name:  "recursive, r",
			Usage: "Should sub directories also be uploaded",
		},
		cli.IntFlag{
			Name:  "parallel, n",
			Value: 3,
			Usage: "Number of simultaneous uploads to perform, defaults to 3",
		},
	},
	Action: uploadDirCLI,
}

func uploadDirCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a directory to upload")
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
	if err := client.UploadDirectory(project, dirPath, recursive, numThreads); err != nil {
		fmt.Println("Directory upload failed:", err)
		os.Exit(1)
	}

	fmt.Println("Directory successfully uploaded.")
}
