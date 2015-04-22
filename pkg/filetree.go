package pkg

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type FileEntry struct {
	Path  string
	Finfo os.FileInfo
}

type ProcessFunc func(done <-chan struct{}, f <-chan FileEntry, result <-chan string)

func walkFiles(done <-chan struct{}, root string) (<-chan FileEntry, <-chan error) {
	filesChan := make(chan FileEntry)
	errChan := make(chan error, 1)
	go func() {
		defer close(filesChan)
		errChan <- filepath.Walk(root, func(path string, finfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !finfo.Mode().IsRegular() {
				return nil
			}
			entry := FileEntry{
				Path:  path,
				Finfo: finfo,
			}
			select {
			case filesChan <- entry:
			case <-done:
				return errors.New("walk cancelled")
			}
			return nil
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
	return results, errChan
}
