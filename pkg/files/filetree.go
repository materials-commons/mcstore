package files

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// A TreeEntry is passed along the channel to the process function.
type TreeEntry struct {
	Path  string
	Finfo os.FileInfo
}

// ProcessFunc is the signature for a function that processes each file entry.
type ProcessFunc func(done <-chan struct{}, f <-chan TreeEntry, result chan<- string)

// IgnoreFunc is a function definition that tells walkFiles when it should
// ignore files and directory paths.
type IgnoreFunc func(path string, finfo os.FileInfo) bool

// walkFiles walks a directory tree. It ignores files and paths with permission errors. All other errors
// will cause it to stop walking the tree. For every path segment it will call ignorePathFn and skip that
// file or sub directory if it returns true. Otherwise all regular files are sent along the filesChan
// for further processing.
func walkFiles(done <-chan struct{}, root string, ignorePathFn IgnoreFunc) (<-chan TreeEntry, <-chan error) {
	// filesChan channel is where we send regular files that aren't ignored.
	filesChan := make(chan TreeEntry)

	// errChan sends errors found while walking the tree.
	errChan := make(chan error, 1)

	// Walk the tree in separate go routine. Send back results and errors in the filesChan and errChan channels.
	go func() {
		defer close(filesChan)
		errChan <- filepath.Walk(root, func(path string, finfo os.FileInfo, err error) error {
			switch {
			case err != nil && os.IsPermission(err):
				// Permission errors are ignored. Just continue walking the tree
				// without processing the file or directory.
				return nil

			case err != nil:
				// All other errors cause walking to stop.
				return err

			case ignorePathFn(path, finfo):
				// if ignorePathFn returns true then skip processing this entry.
				if finfo.IsDir() {
					// If entry is a directory, then skip processing that
					// entry sub tree.
					return filepath.SkipDir
				}
				return nil

			case !finfo.Mode().IsRegular():
				// Only process regular files. This means the following types
				// are not processed: Dir, Symbol Link, Named Pip, Socket,
				// and Device files.
				return nil

			default:
				// This is a regular file, so send it along the fileChan for
				// potential processing.
				entry := TreeEntry{
					Path:  path,
					Finfo: finfo,
				}
				select {
				case filesChan <- entry:
				case <-done:
					// if we receive a message on the done channel, then stop
					// walking the tree.
					return errors.New("walk canceled")
				}
				return nil
			}

		})
	}()
	return filesChan, errChan
}

// PWalk will walk a tree and process files in it in parallel. fn is the process function. "n" go routines will
// be started with fn for processing. The ignorePathFn determines whether a file or directory path should be
// ignored. ignorePathFn can be nil. If it is nil then no files or paths are ignored.
func PWalk(root string, n int, fn ProcessFunc, ignorePathFn IgnoreFunc) (<-chan string, <-chan error) {
	done := make(chan struct{})
	defer close(done)
	// default to never ignoring files and directory paths.
	ignoreFn := neverIgnore
	if ignorePathFn != nil {
		// user supplied a function to test for ignoring entries.
		ignoreFn = ignorePathFn
	}

	filesChan, errChan := walkFiles(done, root, ignoreFn)

	// results holds the results of processing each entry. Its value is
	// ignored and only exists so that PWalk will block until all
	// processing has completed.
	results := make(chan string)

	// Setup a WaitGroup equal to the number of routines processing
	// file entries.
	var wg sync.WaitGroup
	wg.Add(n)

	// Start "n" fn routines to process files. When each one completes
	// processing all entries along its channel signal done.
	for i := 0; i < n; i++ {
		go func() {
			fn(done, filesChan, results)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	for r := range results {
		var _ = r
	}
	return results, errChan
}

// neverIgnore is the default method for checking if an entry should be ignored.
// It never ignores an entry.
func neverIgnore(path string, finfo os.FileInfo) bool {
	return false
}
