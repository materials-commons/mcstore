package mc

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/parnurzeal/gorequest"
)

var projectStatusCommand = cli.Command{
	Name:    "status",
	Aliases: []string{"s", "stat"},
	Usage:   "List status of project",
	Flags: []cli.Flag{
		cli.BoolTFlag{
			Name:  "uploads, u",
			Usage: "Display outstanding upload requests.",
		},
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "Display all status information",
		},
		cli.BoolFlag{
			Name:  "changes, c",
			Usage: "Display all file changes",
		},
	},
	Action: projectStatusCLI,
}

type projectStatusCommandState struct {
	client *gorequest.SuperAgent
}

func projectStatusCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project name.")
		os.Exit(1)
	}

	proj := c.Args()[0]
	s := &projectStatusCommandState{
		client: newGoRequest(),
	}

	switch {
	case c.Bool("all"):
		s.displayStatusAll(proj)
	case c.Bool("uploads"):
		s.displayStatusUploads(proj)
	case c.Bool("changes"):
		s.displayStatusFileChanges(proj)
	}
}

func (s *projectStatusCommandState) displayStatusAll(project string) {

}

func (s *projectStatusCommandState) displayStatusUploads(project string) {

}

func (s *projectStatusCommandState) displayStatusFileChanges(project string) {

}
