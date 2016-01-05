package mccli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var uploadFileCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Upload a file to MaterialsCommons",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project the file is in",
		},
	},
	Action: uploadFileCLI,
}

func uploadFileCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a directory to upload")
		os.Exit(1)
	}

	path := filepath.Clean(c.Args()[0])
	project := c.String("project")

	client := mc.NewClientAPI()

	if err := client.UploadFile(project, path); err != nil {
		fmt.Println("File upload failed:", err)
		os.Exit(1)
	}

	fmt.Println("File successfully uploaded.")
}
