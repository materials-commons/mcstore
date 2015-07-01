package mccli

import "github.com/codegangsta/cli"

var UploadCommand = cli.Command{
	Name:    "upload",
	Aliases: []string{"up", "u"},
	Usage:   "Upload commands",
	Subcommands: []cli.Command{
		uploadProjectCommand,
		uploadFileCommand,
		uploadDirCommand,
	},
}
