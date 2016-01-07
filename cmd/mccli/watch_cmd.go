package mccli

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
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

type projectWatcher struct {
	*mc.ClientAPI
	projectName string
}

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
		p := &projectWatcher{
			ClientAPI:   mc.NewClientAPI(),
			projectName: projectName,
		}
		path := db.Project().Path
		fmt.Printf("Watching project %s located at %s for changes...\n", projectName, path)
		p.watchProject(path, db)
	}
}

func (w *projectWatcher) watchProject(path string, db mc.ProjectDB) {
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
				w.handleFileChangeEvent(event, db)
			case err := <-watcher.ErrorEvents:
				fmt.Println("file events error:", err)
				break FsEventsLoop
			}
		}
		watcher.Close()
	}
}

func (w *projectWatcher) handleFileChangeEvent(event fs.Event, db mc.ProjectDB) {
	switch {
	case event.IsCreate():
		w.handleCreate(event.Name)
	case event.IsDelete():
		// ignore
	case event.IsModify():
		w.handleModify(event.Name)
	case event.IsRename():
		// ignore
	case event.IsAttrib():
		// ignore
	default:
		// ignore
	}
}

func (w *projectWatcher) handleCreate(path string) {
	switch finfo, err := os.Stat(path); {
	case err != nil:
		fmt.Printf("Error stating %s: %s\n", path, err)
	case finfo.IsDir():
		w.dirCreate(path)
	case finfo.Mode().IsRegular():
		w.fileUpload(path)
	}
}

func (w *projectWatcher) dirCreate(path string) {
	if err := w.CreateDirectory(w.projectName, path); err != nil {
		fmt.Printf("Failed to create new directory %s: %s\n", path, err)
	} else {
		fmt.Println("Created new directory: ", path)
	}
}

func (w *projectWatcher) fileUpload(path string) {
	if err := w.UploadFile(w.projectName, path); err != nil {
		fmt.Printf("Failed to upload file %s: %s\n", path, err)
	}
}

func (w *projectWatcher) handleModify(path string) {
	if finfo, err := os.Stat(path); err != nil {
		fmt.Printf("Error getting file info for %s: %s\n", path, err)
	} else if finfo.Mode().IsRegular() {
		w.fileUpload(path)
	}
}
