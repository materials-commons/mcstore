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

// PWalker implements a parallel walker. It holds the options and processing
// methods to use during the walk.
type PWalker struct {
	// NumParallel is the number of go routines to start for processing.
	NumParallel int

	// ProcessFn is the function to use to process entries.
	ProcessFn ProcessFunc

	// IgnoreFn is the method to test if an entry should be ignored. If
	// IgnoreFn is nil then all entries will be process.
	IgnoreFn IgnoreFunc

	// ProcessDirs tells the walker whether directory entries should also
	// be passed to ProcessFn. It defaults to false, so by default directories
	// are not passed to ProcessFn.
	ProcessDirs bool
}

// walkFiles walks a directory tree. It ignores files and paths with permission errors. All other errors
// will cause it to stop walking the tree. For every path segment it will call ignorePathFn and skip that
// file or sub directory if it returns true. Otherwise all regular files are sent along the filesChan
// for further processing.
func (p *PWalker) walkFiles(done <-chan struct{}, root string) (<-chan TreeEntry, <-chan error) {
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

			case p.IgnoreFn(path, finfo):
				// if ignorePathFn returns true then skip processing this entry.
				if finfo.IsDir() {
					// If entry is a directory, then skip processing that
					// entry sub tree.
					return filepath.SkipDir
				}
				return nil

			case !finfo.Mode().IsRegular() && !finfo.IsDir():
				// Only process regular files. This means the following types
				// are not processed: Symbol Link, Named Pip, Socket,
				// and Device files.
				//
				// Directories are a special case and are handled in the default
				// clause, where regulars files are also handled.
				return nil

			default:
				// Directories aren't being processed, so skip this entry.
				if finfo.IsDir() && !p.ProcessDirs {
					return nil
				}

				// This is an entry we can process, so send it along the fileChan
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
func (p *PWalker) PWalk(root string) (<-chan string, <-chan error) {
	done := make(chan struct{})
	defer close(done)

	if p.IgnoreFn == nil {
		p.IgnoreFn = neverIgnore
	}

	if p.NumParallel < 1 {
		p.NumParallel = 3
	}

	filesChan, errChan := p.walkFiles(done, root)

	// results holds the results of processing each entry. Its value is
	// ignored and only exists so that PWalk will block until all
	// processing has completed.
	results := make(chan string)

	// Setup a WaitGroup equal to the number of routines processing
	// file entries.
	var wg sync.WaitGroup
	wg.Add(p.NumParallel)

	// Start "n" fn routines to process files. When each one completes
	// processing all entries along its channel signal done.
	for i := 0; i < p.NumParallel; i++ {
		go func() {
			p.ProcessFn(done, filesChan, results)
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
func neverIgnore(_ string, _ os.FileInfo) bool {
	return false
}
