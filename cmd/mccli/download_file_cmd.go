package mccli

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var downloadFileCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Download a project file",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project to download the file from",
		},
	},
	Action: downloadFileCLI,
}

func downloadFileCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a file to download.")
		os.Exit(1)
	}

	path := filepath.Clean(c.Args()[0])
	project := c.String("project")
	client := mc.NewClientAPI()

	if err := client.DownloadFile(project, path); err != nil {
		fmt.Println("File download failed:", err)
		os.Exit(1)
	}

	fmt.Println("File successfully downloaded.")
}
