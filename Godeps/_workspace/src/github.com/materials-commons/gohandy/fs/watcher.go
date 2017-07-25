/*
Package fs implements a recursive file system watcher. This code is based on code from
https://github.com/gophertown/looper/blob/master/watch.go.
*/
package fs

import (
	"errors"
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"path/filepath"
)

// Event is a structure that wraps the fsnotify.FileEvent. The reason
// we don't expose fsnotify.FileEvent directly is to allow for future
// expansion of state information in the Event struct.
type Event struct {
	*fsnotify.FileEvent
}

// RecursiveWatcher represents the file system watcher and
// communication channels.
type RecursiveWatcher struct {
	*fsnotify.Watcher            // File system watcher
	Events            chan Event // The channel to send events on
	ErrorEvents       chan error
}

// NewRecursiveWatcher creates a new file system watcher for path. It walks the directory
// tree at path adding each directory to the watcher. When the user creates a new directory
// it is also added to the list of directories to watch.
func NewRecursiveWatcher(path string) (*RecursiveWatcher, error) {
	directories := subdirs(path)
	if len(directories) == 0 {
		return nil, errors.New("no directories to watch")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	recursiveWatcher := &RecursiveWatcher{Watcher: watcher}
	recursiveWatcher.Events = make(chan Event, 10)

	for _, dir := range directories {
		recursiveWatcher.addDirectory(dir)
	}

	return recursiveWatcher, nil
}

// addDirectory adds a new directory to the file system watcher. It is automatically
// called when the watcher detects a directory create event.
func (watcher *RecursiveWatcher) addDirectory(dir string) {
	err := watcher.WatchFlags(dir, fsnotify.FSN_ALL)
	if err != nil {
		log.Println("Error watching directory: ", dir, err)
	}
}

// Start starts monitoring for file system events and sends the
// events on the Events channel. It also handles directory create
// events by adding the newly created directory to the list of
// directories to monitor.
func (watcher *RecursiveWatcher) Start() {
	go func() {
		for {
			select {
			case event := <-watcher.Event:
				watcher.handleEvent(event)
			case err := <-watcher.Error:
				watcher.ErrorEvents <- err
				log.Println("error:", err)
			}
		}
	}()
}

// handleEvent performs the actual task of adding directories newly created
// directories to be monitored and sending events on the Events channel.
func (watcher *RecursiveWatcher) handleEvent(event *fsnotify.FileEvent) {
	if event.IsCreate() {
		watcher.handleCreate(event)
	}

	e := Event{
		FileEvent: event,
	}

	watcher.Events <- e
}

// handleCreate checks if the created item is a directory and if so
// sets that directory up for monitoring.
func (watcher *RecursiveWatcher) handleCreate(event *fsnotify.FileEvent) {
	finfo, err := os.Stat(event.Name)
	if err != nil {
		log.Printf("Error on stat for %s: %s\n", event.Name, err.Error())
	} else if finfo.IsDir() {
		watcher.addDirectory(event.Name)
	}
}

// subdirs walks a directory creating a list of all subdirectories. It ignores
// hidden directories (directories starting with a '.' "dot")
func subdirs(path string) (paths []string) {
	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			hidden := filepath.HasPrefix(name, ".") && name != "." && name != ".."
			if hidden {
				return filepath.SkipDir
			}
			paths = append(paths, subpath)
		}
		return nil
	})
	return paths
}
