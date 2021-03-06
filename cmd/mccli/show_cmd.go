package mccli

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
)

var ShowCommand = cli.Command{
	Name:    "show",
	Aliases: []string{"sh"},
	Usage:   "Show commands",
	Subcommands: []cli.Command{
		showConfigCommand,
		showProjectCommand,
	},
}

var showConfigCommand = cli.Command{
	Name:    "config",
	Aliases: []string{"conf", "c"},
	Usage:   "Show configuration",
	Action:  showConfigCLI,
}

func showConfigCLI(c *cli.Context) {
	fmt.Println("apikey:", config.GetString("apikey"))
	fmt.Println("mcurl:", mcstoreapi.MCUrl())
	fmt.Println("mclogging:", config.GetString("mclogging"))
}

var showProjectCommand = cli.Command{
	Name:    "project",
	Aliases: []string{"proj", "p"},
	Usage:   "Show information on project",
	Action:  showProjectCLI,
}

const uploadTimeFormat = "Mon, 02 Jan 2006 at 3:04PM"

func showProjectCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project.")
		os.Exit(1)
	}

	projectName := c.Args()[0]

	if projectDB, err := mc.ProjectOpener.OpenProjectDB(projectName); err != nil {
		fmt.Println("Unknown project:", projectName)
		os.Exit(1)
	} else {
		project := projectDB.Project()
		fmt.Println("Project:", project.Name)
		fmt.Println("Path   :", project.Path)
		fmt.Println("ID     :", project.ProjectID)
		filepath.Walk(project.Path, func(path string, finfo os.FileInfo, err error) error {
			switch {
			case err != nil:
				// nothing to do
			case finfo.IsDir():
				if files.IgnoreDotAndTempFiles(path, finfo) {
					fmt.Printf("\nDirectory: %s is ignored\n", path)
					return filepath.SkipDir
				}
				showDirStatus(path, finfo, projectDB)
			case finfo.Mode().IsRegular():
				showFileStatus(path, finfo, projectDB)
			}
			return nil
		})
	}
}

func showDirStatus(path string, finfo os.FileInfo, projectDB mc.ProjectDB) {
	if dir, err := projectDB.FindDirectory(path); err != nil {
		fmt.Printf("\nDirectory: %s is new\n", path)
	} else {
		var _ = dir
		fmt.Printf("\nDirectory:%s\n", path)
	}
}

func showFileStatus(path string, finfo os.FileInfo, projectDB mc.ProjectDB) {
	if files.IgnoreDotAndTempFiles(path, finfo) {
		fmt.Printf("  File: %s is ignored\n", finfo.Name())
	} else {
		fileDir := filepath.Dir(path)
		if dir, err := projectDB.FindDirectory(fileDir); err != nil {
			fmt.Printf("  File: %s is new\n", finfo.Name())
		} else {
			if f, err := projectDB.FindFile(finfo.Name(), dir.ID); err != nil {
				fmt.Printf("  File: %s is new\n", finfo.Name())
			} else if finfo.ModTime().Unix() > f.MTime.Unix() {
				fmt.Printf("  File: %s has changed, last uploaded on %s\n",
					finfo.Name(), f.LastUpload.Format(uploadTimeFormat))
			} else {
				fmt.Printf("  File: %s was uploaded on %s\n",
					finfo.Name(), f.LastUpload.Format(uploadTimeFormat))
			}
		}
	}
}
