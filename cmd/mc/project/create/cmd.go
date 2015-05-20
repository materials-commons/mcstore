package create

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/opts"
	"github.com/materials-commons/mcstore/cmd/pkg/project"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/parnurzeal/gorequest"
)

type createCommandArgs struct {
	projectName   string
	projectID     string
	directoryPath string
	n             int
}

var (
	// Command contains the arguments and function for the cli project create command.
	Command = cli.Command{
		Name:    "create",
		Aliases: []string{"cr", "c"},
		Usage:   "Create a new project",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "directory, dir, d",
				Usage: "The base directory for the project",
			},
			cli.IntFlag{
				Name:  "parallel, n",
				Value: 3,
				Usage: "Number of simultaneous uploads to perform, defaults to 3",
			},
		},
		Action: createCLI,
	}

	args createCommandArgs
)

// createCLI implements the project create command.
func createCLI(c *cli.Context) {
	if err := validateArgs(c); err != nil {
		fmt.Println("Invalid arguments:", err)
		os.Exit(1)
	}

	proj, err := project.Create(args.directoryPath, args.projectName, args.projectID)
	if err != nil {
		fmt.Println("Unable to create project:", err)
		os.Exit(1)
	}

	fmt.Println("Indexing project...")
	indexProject(args.directoryPath, proj)
	fmt.Println("Done.")
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
	if err := validateDirectoryPath(args.directoryPath); err != nil {
		return err
	}

	args.n = opts.GetNumThreads(c)

	return nil
}

// validateProject ensures that the projectID exists and the user has access.
func createProject(projectName string) error {

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

// indexProject walks the directory tree and indexes each of the files found. Indexing
// can be performed in parallel.
func indexProject(path string, proj *project.MCProject) error {
	fn := func(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
		indexEntries(proj, done, entries, result)
	}
	walker := &files.PWalker{
		NumParallel: args.n,
		ProcessFn:   fn,
		ProcessDirs: true,
	}
	walker.PWalk(args.directoryPath)
	return nil
}

// indexEntries processes the entries sent along the entries channel. It also
// processes done channel events by exiting the go routine.
func indexEntries(proj *project.MCProject, done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
	client := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	for entry := range entries {
		select {
		case result <- indexEntry(proj, entry, client):
		case <-done:
			return
		}
	}
}

func indexEntry(proj *project.MCProject, entry files.TreeEntry, client *gorequest.SuperAgent) string {
	switch {
	case entry.Finfo.IsDir():

	default:
	}
	return ""
}
