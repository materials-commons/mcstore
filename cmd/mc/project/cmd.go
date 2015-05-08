package project

import (
	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/mc/project/upload"
)

// Command contains the project command and all its sub commands.
var Command = cli.Command{
	Name:    "project",
	Aliases: []string{"proj", "p"},
	Usage:   "Project commands",
	Subcommands: []cli.Command{
		upload.Command,
	},
}
