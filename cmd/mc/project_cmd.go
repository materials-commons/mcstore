package mc

import "github.com/codegangsta/cli"

// Command contains the project command and all its sub commands.
var ProjectCommand = cli.Command{
	Name:    "project",
	Aliases: []string{"proj", "p"},
	Usage:   "Project commands",
	Subcommands: []cli.Command{
		projectUploadCommand,
		projectCreateCommand,
		projectStatusCommand,
	},
}
