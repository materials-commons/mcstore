package mccli

import "github.com/codegangsta/cli"

var DownloadCommand = cli.Command{
	Name:    "download",
	Aliases: []string{"down", "d"},
	Usage:   "Downloads files, directories or projects",
	Subcommands: []cli.Command{
		downloadProjectCommand,
		downloadFileCommand,
		downloadDirCommand,
	},
}
