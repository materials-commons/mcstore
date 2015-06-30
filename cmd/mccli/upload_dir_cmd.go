package mccli

import "github.com/codegangsta/cli"

var uploadDirCommand = cli.Command{
	Name:    "file",
	Aliases: []string{"f"},
	Usage:   "Upload a file to MaterialsCommons",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "project, proj, p",
			Usage: "The project the file is in",
		},
		//		cli.BoolFlag{
		//			Name:  "recursive, r",
		//			Usage: "Should sub directories also be uploaded",
		//		},
	},
	Action: uploadDirCLI,
}

func uploadDirCLI(c *cli.Context) {

}
