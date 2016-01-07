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
				fmt.Println("file events error:", err)
				break FsEventsLoop
			}
		}
		watcher.Close()
	}
}

func handleFileChangeEvent(event *fsnotify.FileEvent, db mc.ProjectDB) {
	switch {
	case event.IsCreate():
		handleCreate(event.Name)
	case event.IsDelete():
		// ignore
	case event.IsModify():
		handleModify(event.Name)
	case event.IsRename():
		// ignore
	case event.IsAttrib():
		// ignore
	default:
		// ignore
	}
}

func handleCreate(path string) {
	switch finfo, err := os.Stat(path); {
	case err != nil:
		fmt.Printf("Error stating %s: %s\n", path, err)
	case finfo.IsDir():
		handleDirCreate(path, finfo)
	case finfo.Mode().IsRegular():
		handleFileCreate(path, finfo)
	}
}

func handleDirCreate(path string, finfo os.FileInfo) {
	// create new directory
}

func handleFileCreate(path string, finfo os.FileInfo) {
	// upload new file
}

func handleModify(path string) {
	if finfo, err := os.Stat(path); err != nil {
		fmt.Printf("Error stating %s: %s\n", path, err)
	} else if finfo.Mode().IsRegular() {
		handleFileModify(path, finfo)
	}
}

func handleFileModify(path string, finfo os.FileInfo) {
	// upload changed file
}
