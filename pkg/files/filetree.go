package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type TreeEntry struct {
	Path  string
	Finfo os.FileInfo
}

type ProcessFunc func(done <-chan struct{}, f <-chan TreeEntry, result chan<- string)

func walkFiles(done <-chan struct{}, root string) (<-chan TreeEntry, <-chan error) {
	filesChan := make(chan TreeEntry)
	errChan := make(chan error, 1)
	go func() {
		defer close(filesChan)
		errChan <- filepath.Walk(root, func(path string, finfo os.FileInfo, err error) error {
			switch {
			case err != nil && os.IsPermission(err):
				//fmt.Println("Got permission denied, continuing", path)
				return nil
			case err != nil:
				fmt.Println("walk got error", err)
				return err
			case !finfo.Mode().IsRegular():
				return nil
			default:
				entry := TreeEntry{
					Path:  path,
					Finfo: finfo,
				}
				select {
				case filesChan <- entry:
				case <-done:
					return errors.New("walk cancelled")
				}
				return nil
			}

		})
	}()
	return filesChan, errChan
}

func PWalk(root string, n int, fn ProcessFunc) (<-chan string, <-chan error) {
	done := make(chan struct{})
	defer close(done)
	filesChan, errChan := walkFiles(done, root)
	results := make(chan string)
	var wg sync.WaitGroup
	wg.Add(n)
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
