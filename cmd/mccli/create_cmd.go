package mccli

import "github.com/codegangsta/cli"

var CreateCommand = cli.Command{
	Name:    "create",
	Aliases: []string{"cr", "c"},
	Usage:   "Create commands",
	Subcommands: []cli.Command{
		createProjectCommand,
	},
}
