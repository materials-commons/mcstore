package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

// createCommandArgs holds values that won't change and are
// needed during the upload process.
type projectCreateCommandArgs struct {
	projectName   string
	projectID     string
	directoryPath string
	n             int
}

var (
	// Command contains the arguments and function for the cli project create command.
	projectCreateCommand = cli.Command{
		Name:    "create",
		Aliases: []string{"cr", "c"},
		Usage:   "Create a new project",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "directory, dir, d",
				Usage: "The base directory for the project",
			},
		},
		Action: projectCreateCLI,
	}

	// args contains global values, including arguments from the cli
	// that are needed to create and upload a project.
	args projectCreateCommandArgs
)

// createCLI implements the project create command.
func projectCreateCLI(c *cli.Context) {
	if err := validateArgs(c); err != nil {
		fmt.Println("Invalid arguments:", err)
		os.Exit(1)
	}

	p := mc.ProjectDBSpec{
		Name:      args.projectName,
		Path:      args.directoryPath,
		ProjectID: args.projectID,
	}
	_, err := mc.ProjectOpener.CreateProjectDB(p)
	if err != nil {
		fmt.Println("Unable to create project:", err)
		os.Exit(1)
	}
}

// validate will validate the command line arguments. It will print a message
// and exit if there is a bad argument.
func validateArgs(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return fmt.Errorf("You must supply a name for the project.")
	}
	args.projectName = c.Args()[0]

	if err := createProject(args.projectName); err != nil {
		return err
	}

	args.directoryPath = c.String("directory")
	return validateDirectoryPath(args.directoryPath)
}

// createProject creates the new project for the user.
func createProject(projectName string) error {
	//	req := mcstore.CreateProjectRequest{
	//		Name:         projectName,
	//		MustNotExist: true,
	//	}
	//
	//	var resp mcstore.CreateProjectResponse
	//	client := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	//	if err := sendRequest(client, "/projects", req, &resp); err != nil {
	//		fmt.Println("Unable to create project:", err)
	//		return err
	//	}
	//
	//	args.projectID = resp.ProjectID
	return nil
}

// validateDirectoryPath checks that the given directory path exists.
func validateDirectoryPath(path string) error {
	if path == "" {
		return fmt.Errorf("You must specify a local directory path where the project files are located.")
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("Invalid directory: %s", path)
	}
	return nil
}
