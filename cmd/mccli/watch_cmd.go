package mccli

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/howeyc/fsnotify"
	"github.com/materials-commons/gohandy/fs"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
)

var (
	WatchCommand = cli.Command{
		Name:    "watch",
		Aliases: []string{"w"},
		Usage:   "Watch commands",
		Subcommands: []cli.Command{
			watchProjectCommand,
		},
	}

	watchProjectCommand = cli.Command{
		Name:    "project",
		Aliases: []string{"proj", "p"},
		Usage:   "Watch a project for file changes",
		Action:  watchProjectCLI,
	}
)

func watchProjectCLI(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("You must specify a project.")
		os.Exit(1)
	}

	projectName := c.Args()[0]

	if db, err := mc.ProjectOpener.OpenProjectDB(projectName); err != nil {
		fmt.Println("Unknown project:", projectName)
		os.Exit(1)
	} else {
		project := db.Project()
		path := project.Path
		watchProject(path, db)
	}
}

func watchProject(path string, db mc.ProjectDB) {
	for {
		watcher, err := fs.NewRecursiveWatcher(path)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}
		watcher.Start()

	FsEventsLoop:
		for {
			select {
			case event := <-watcher.Events:
				handleFileChangeEvent(event, db)
			case err := <-watcher.ErrorEvents:
				fmt.Println("error:", err)
				break FsEventsLoop
			}
		}
		watcher.Close()
	}
}

func handleFileChangeEvent(event *fsnotify.FileEvent, db mc.ProjectDB) {
	// check type of event, then check if file or dir. If dir and new then create.
	// If file and new or changed, then upload. Ignore delete and rename for the moment.
}
