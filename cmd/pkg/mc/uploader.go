package mc

import (
	"time"

	"crypto/md5"
	"io"
	"math"
	"os"

	"path/filepath"

	"fmt"

	"sync"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/gohandy/with"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
)

// projectUploader holds the starting state for a project upload. It controls
// how many threads process requests.
type projectUploader struct {
	db         ProjectDB
	numThreads int
}

// uploadProject will upload the project. It starts a parallel walker to process entries.
func (p *projectUploader) uploadProject() error {
	db := p.db
	project := db.Project()

	// create a func to process entries.
	fn := func(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
		uploader := newUploader(p.db, project)
		uploader.retrier.RetryCount = 1
		uploader.uploadEntries(done, entries, result)
	}

	walker := files.PWalker{
		NumParallel: 1,
		ProcessFn:   fn,
		ProcessDirs: true,
	}

	_, errc := walker.PWalk(project.Path)
	err := <-errc
	return err
}

// uploadFile will upload a single file.
func (p *projectUploader) uploadFile(path string) error {
	uploader := newUploader(p.db, p.db.Project())
	finfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	treeEntry := files.TreeEntry{
		Finfo: finfo,
		Path:  path,
	}

	uploader.handleFileEntry(treeEntry)
	return nil
}

// uploadDirectory will upload a single directory. It will ignore sub directories when ignoreSubDirectories
// is true.
func (p *projectUploader) uploadDirectory(path string, recursive bool) error {
	// create a custom ignore function that will ignore sub directories if recursive is false.
	ignoreFunc := func(pathEntry string, fileInfo os.FileInfo) bool {
		if fileInfo.IsDir() && !recursive {
			// Special case pathEntry is the dir we are uploading. In that case return
			// false (do not ignore)
			if path == pathEntry {
				return false
			}
			return true
		}
		return files.IgnoreDotAndTempFiles(pathEntry, fileInfo)
	}
	db := p.db
	project := db.Project()

	// create a func to process entries.
	fn := func(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
		uploader := newUploader(p.db, project)
		uploader.retrier.RetryCount = 1
		uploader.uploadEntries(done, entries, result)
	}

	walker := files.PWalker{
		NumParallel: 1,
		ProcessFn:   fn,
		ProcessDirs: true,
		IgnoreFn:    ignoreFunc,
	}

	_, errc := walker.PWalk(path)
	err := <-errc
	return err
}

// uploader holds the state for one upload thread.
type uploader struct {
	db        ProjectDB
	serverAPI *mcstoreapi.ServerAPI
	project   *Project
	retrier   with.Retrier
}

// newUploader creates a new uploader. It creates a clone of the database.
func newUploader(db ProjectDB, project *Project) *uploader {
	return &uploader{
		db:        db.Clone(),
		project:   project,
		serverAPI: mcstoreapi.NewServerAPI(),
	}
}

// uploadEntries waits on the entries and done channels. When  entry comes in on
// the entries channel it will process it. If an something comes in on the done
// channel then it will stop processing and return from the go routine.
func (u *uploader) uploadEntries(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
	for entry := range entries {
		select {
		case result <- u.uploadEntry(entry):
		case <-done:
			return
		}
	}
}

// uploadEntry identifies the type of entry it is processing (file or directory) and
// processes it appropriately.
func (u *uploader) uploadEntry(entry files.TreeEntry) string {
	switch {
	case entry.Finfo.IsDir():
		u.handleDirEntry(entry)
	default:
		u.handleFileEntry(entry)
	}
	return ""
}

// handleDirEntry handles processing of directories. It checks if the directory already exists in the
// local database. If it doesn't it will create the directory on the server and insert the directory
// into its local database.
func (u *uploader) handleDirEntry(entry files.TreeEntry) {
	path := filepath.ToSlash(entry.Path)

	_, err := u.db.FindDirectory(path)
	switch {
	case err == app.ErrNotFound:
		u.createDirectory(entry)
	case err != nil:
		app.Log.Panicf("Local database returned err, panic!: %s", err)
	default:
		// directory already known nothing to do
		return
	}
}

// createDirectory creates a new directory entry on the server.
func (u *uploader) createDirectory(entry files.TreeEntry) {
	dirPath := filepath.ToSlash(entry.Path)
	req := mcstoreapi.DirectoryRequest{
		ProjectName: u.project.Name,
		ProjectID:   u.project.ProjectID,
		Path:        dirPath,
	}
	dirID, _ := u.getDirectory(req)
	dir := &Directory{
		DirectoryID: dirID,
		Path:        dirPath,
	}
	if _, err := u.db.InsertDirectory(dir); err != nil {
		app.Log.Panicf("Local database returned err, panic!: %s", err)
	}
}

// getDirectoryWithRetry will make the server API GetDirectory call. If it fails
// it will retry (dependent on retry settings).
func (u *uploader) getDirectory(req mcstoreapi.DirectoryRequest) (string, error) {
	var (
		dirID string
		err   error
	)
	if dirID, err = u.serverAPI.GetDirectory(req); err != nil {
		app.Log.Errorf("Failed creating directory: %#v/%s", req, err)
		return "", err
	}
	return dirID, nil
}

// handleFileEntry will handle processing file entries. It will check if the file
// has already been uploaded. If it hasn't it will upload the file to the server.
func (u *uploader) handleFileEntry(entry files.TreeEntry) {
	if dir := u.getDirByPath(filepath.Dir(entry.Path)); dir == nil {
		app.Log.Exitf("Should have found dir %s", filepath.Dir(entry.Path))
	} else {
		file := u.getFileByName(entry.Finfo.Name(), dir.ID)
		switch {
		case file == nil:
			fmt.Println("Uploading new file:", entry.Path)
			u.uploadFile(entry, file, dir)
		case entry.Finfo.ModTime().Unix() > file.MTime.Unix():
			fmt.Println("Uploading changed file:", entry.Path)
			u.uploadFile(entry, file, dir)
		default:
			fmt.Println("File already uploaded:", entry.Path)
			// nothing to do
		}
	}
}

// getDirByPath looks up a path in the local database. It normalizes
// the file separator.
func (u *uploader) getDirByPath(path string) *Directory {
	dir, err := u.db.FindDirectory(path)
	switch {
	case err == app.ErrNotFound:
		return nil
	case err != nil:
		app.Log.Panicf("Local database returned err, panic!: %s", err)
		return nil
	default:
		// directory already known nothing to do
		return dir
	}
}

// getFileByName looks up a file by name and directory id in the local database.
func (u *uploader) getFileByName(name string, dirID int64) *File {
	f, err := u.db.FindFile(name, dirID)
	switch {
	case err == app.ErrNotFound:
		return nil
	case err != nil:
		app.Log.Panicf("Local database returned err, panic!: %s", err)
		return nil
	default:
		return f
	}
}

// uploadFile will upload the file blocks to the server. It will create an upload request
// and then start uploading the blocks. When it completes it will update the local database
// state for the file.
func (u *uploader) uploadFile(entry files.TreeEntry, file *File, dir *Directory) {
	uploadResponse, checksum := u.getUploadResponse(dir.DirectoryID, entry)
	requestID := uploadResponse.RequestID

	var (
		_ = checksum

		// TODO: do something with the starting block (its ignored for now)
		n          int
		err        error
		uploadErr  error
		mutex      sync.Mutex
		uploadResp *mcstoreapi.UploadChunkResponse
		wg         sync.WaitGroup
	)
	uploadChan := make(chan *flow.Request)
	chunkNumber := 1

	done := make(chan struct{})

	uploadFunc := func(doneChan <-chan struct{}, c <-chan *flow.Request) {
		defer wg.Done()
		for req := range c {
			select {
			case <-doneChan:
				return

			default:
				resp, err := u.sendFlowData(req)
				if err != nil {
					mutex.Lock()
					uploadErr = err
					mutex.Unlock()
				} else {
					if resp.Done {
						mutex.Lock()
						uploadResp = resp
						mutex.Unlock()
					}
				}
			}
		}
	}

	wg.Add(5)

	for i := 0; i < 5; i++ {
		go uploadFunc(done, uploadChan)
	}

	f, _ := os.Open(entry.Path)
	defer f.Close()
	buf := make([]byte, 1024*1024)
	totalChunks := numChunks(entry.Finfo.Size())
	//var uploadResp *mcstoreapi.UploadChunkResponse
	for {
		n, err = f.Read(buf)
		if n != 0 {
			// send bytes
			req := &flow.Request{
				FlowChunkNumber:  int32(chunkNumber),
				FlowTotalChunks:  totalChunks,
				FlowChunkSize:    int32(n),
				FlowTotalSize:    entry.Finfo.Size(),
				FlowIdentifier:   requestID,
				FlowFileName:     entry.Finfo.Name(),
				FlowRelativePath: "",
				ProjectID:        u.project.ProjectID,
				DirectoryID:      dir.DirectoryID,
				Chunk:            buf[:n],
			}
			uploadChan <- req
			//			uploadResp, _ = u.sendFlowData(req)
			//			if uploadResp.Done {
			//				break
			//			}
			chunkNumber++
		}
		if err != nil {
			break
		}
	}

	close(done)
	wg.Wait()

	// ******************************* ADD THIS *********************
	// Need to wait on all go routines to finish before proceeding
	// **************************************************************

	if uploadResp == nil {
		app.Log.Errorf("uploadResp not done %#v\n", uploadResp)
		return
	}

	if err != nil && err != io.EOF {
		app.Log.Errorf("Unable to complete read on file for upload: %s", entry.Path)
	} else {
		// done, add or update the entry in the database.
		if file == nil {
			// create new entry
			newFile := File{
				FileID:     uploadResp.FileID,
				Name:       entry.Finfo.Name(),
				Checksum:   checksum,
				Size:       entry.Finfo.Size(),
				MTime:      entry.Finfo.ModTime(),
				LastUpload: time.Now(),
				Directory:  dir.ID,
			}
			u.db.InsertFile(&newFile)
		} else {
			// update existing entry
			file.MTime = entry.Finfo.ModTime()
			file.LastUpload = time.Now()
			if file.Checksum != checksum {
				// Existing file uploaded but we created a new version on the server.
				// We could get here if a previous upload did not complete.
				file.Checksum = checksum
				file.FileID = uploadResp.FileID
				file.Size = entry.Finfo.Size()
			}
			u.db.UpdateFile(file)
		}
	}
}

// getUploadResponse sends an upload request to the server and gets the response.
func (u *uploader) getUploadResponse(directoryID string, entry files.TreeEntry) (*mcstoreapi.CreateUploadResponse, string) {
	checksum, _ := file.HashStr(md5.New(), entry.Path)
	chunkSize := int32(1024 * 1024)
	uploadReq := mcstoreapi.CreateUploadRequest{
		ProjectID:   u.project.ProjectID,
		DirectoryID: directoryID,
		FileName:    entry.Finfo.Name(),
		FileSize:    entry.Finfo.Size(),
		ChunkSize:   chunkSize,
		FileMTime:   entry.Finfo.ModTime().Format(time.RFC1123),
		Checksum:    checksum,
	}
	resp, _ := u.createUploadRequest(uploadReq)
	return resp, checksum
}

// createUploadRequestWithRetry will make the server API CreateUploadRequest call. If it fails
// it will retry (dependent on retry settings).
func (u *uploader) createUploadRequest(uploadReq mcstoreapi.CreateUploadRequest) (*mcstoreapi.CreateUploadResponse, error) {
	var (
		resp *mcstoreapi.CreateUploadResponse
		err  error
	)
	if resp, err = u.serverAPI.CreateUpload(uploadReq); err != nil {
		return nil, err
	}

	return resp, nil
}

// numChunks determines the number of chunks that will be sent to the server.
func numChunks(size int64) int32 {
	d := float64(size) / float64(1024*1024)
	n := int(math.Ceil(d))
	return int32(n)
}

// sendFlowDataWithRetry will make the server API SendFlowData call. If it fails
// it will retry (dependent on retry settings).
func (u *uploader) sendFlowData(req *flow.Request) (*mcstoreapi.UploadChunkResponse, error) {
	var (
		resp *mcstoreapi.UploadChunkResponse
		err  error
	)
	if resp, err = u.serverAPI.SendFlowData(req); err != nil {
		return nil, err
	}
	return resp, nil
}
