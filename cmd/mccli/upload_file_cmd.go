package mccli

import "github.com/codegangsta/cli"

var uploadFileCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Upload a file to MaterialsCommons",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project the file is in",
		},
		cli.StringFlag{
			Name:  "directory, dir, d",
			Usage: "The directory the file is in",
		},
	},
	Action: uploadFileCLI,
}

func uploadFileCLI(c *cli.Context) {

}
