package mc

import (
	"fmt"
	"os"

	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/cmd/pkg/client"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/server/mcstore"
	"github.com/olekukonko/tablewriter"
	"github.com/parnurzeal/gorequest"
)

// projectStatusCommand describes the project status command
// for the cli.
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

// projectStatusCommandState contains all the state information
// needed for the project status command.
type projectStatusCmd struct {
	client *gorequest.SuperAgent
}

// projectStatusCLI implements the project status command.
func projectStatusCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project name.")
		os.Exit(1)
	}

	project := c.Args()[0]
	s := &projectStatusCmd{
		client: client.NewGoRequest(),
	}

	projectID, _ := s.projectName2ID(project)

	switch {
	case c.Bool("all"):
		s.displayStatusAll(projectID)
	case c.Bool("uploads"):
		s.displayStatusUploads(projectID)
	case c.Bool("changes"):
		s.displayStatusFileChanges(projectID)
	}
}

func (s *projectStatusCmd) projectName2ID(projectName string) (string, error) {
	return "", nil
}

// displayStatusAll displays all outstanding uploads and file changes
// for the project.
func (s *projectStatusCmd) displayStatusAll(projectID string) {
	s.displayStatusUploads(projectID)
	s.displayStatusFileChanges(projectID)
}

// displayStatusUploads display all the outstanding uploads for the project.
func (s *projectStatusCmd) displayStatusUploads(projectID string) {
	if uploads, err := s.getUploads(projectID); err != nil {
		fmt.Println("Failed retrieving uploads for project: %s", err)
	} else {
		if len(uploads) == 0 {
			fmt.Printf("There are no upload requests for project %s\n", projectID)
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Date", "File", "Size", "Host"})
			for _, entry := range uploads {
				data := []string{
					entry.RequestID,
					entry.Birthtime.Format(time.RFC822),
					entry.FileName,
					strconv.Itoa(entry.Size),
					entry.Host,
				}
				table.Append(data)
			}
			table.SetBorder(false)
			table.Render()
		}
	}
}

// getUploads queries the server for the uploads for the project.
func (s *projectStatusCmd) getUploads(projectID string) ([]mcstore.UploadEntry, error) {
	config.Set("apikey", "test")
	r, body, errs := s.client.Get(app.MCApi.APIUrl("/upload/test")).End()
	if err := app.MCApi.APIError(r, errs); err != nil {
		return nil, err
	}

	var uploads []mcstore.UploadEntry
	app.MCApi.ToJSON(body, &uploads)
	return uploads, nil
}

// displayStatusFileChanges shows all the files that have changed on the server
// that are not on your local project.
func (s *projectStatusCmd) displayStatusFileChanges(projecID string) {
	fmt.Println("file changes not yet implemented")
	os.Exit(1)
}
