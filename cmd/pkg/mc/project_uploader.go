package mc

import (
	"math/rand"
	"time"

	"crypto/md5"
	"io"
	"math"
	"os"

	"path/filepath"

	"fmt"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstore"
)

var _ = fmt.Println

// projectUploader holds the starting state for a project upload. It controls
// how many threads process requests.
type projectUploader struct {
	db         ProjectDB
	numThreads int
}

// upload will upload the project. It starts a parallel walker to process entries.
func (p *projectUploader) upload() error {
	db := p.db
	project := db.Project()

	// create a func to process entries. Since we have some state to close over we create
	// a closure here to call.
	fn := func(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
		u := newUploader(p.db, project)
		u.uploadEntries(done, entries, result)
	}

	walker := files.PWalker{
		NumParallel: p.numThreads,
		ProcessFn:   fn,
		ProcessDirs: true,
	}

	_, errc := walker.PWalk(project.Path)
	err := <-errc
	return err
}

// uploader holds the state for one upload thread.
type uploader struct {
	db         ProjectDB
	serverAPI  *mcstore.ServerAPI
	project    *Project
	minWait    int
	maxWait    int
	retryCount int
}

// defaultMinWaitBeforeRetry is the default minimum wait time before
// a request is retried.
const defaultMinWaitBeforeRetry = 100

// defaultMaxWaitBeforeRetry is the default max wait time before
// a request is retried.
const defaultMaxWaitBeforeRetry = 5000

// retryForever means we should retry requests forever. If requests shouldn't
// be retried forever then the upload.retryCount should be set to a positive
// number.
const retryForever = -1

// newUploader creates a new uploader. It creates a clone of the database,
// sets minWait to defaultMinWaitBeforeRetry, maxWait to defaultMaxWaitBeforeRetry
// and retryCount to retryForever.
func newUploader(db ProjectDB, project *Project) *uploader {
	return &uploader{
		db:         db.Clone(),
		project:    project,
		serverAPI:  mcstore.NewServerAPI(),
		minWait:    defaultMinWaitBeforeRetry,
		maxWait:    defaultMaxWaitBeforeRetry,
		retryCount: retryForever,
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
	path := filepath.ToSlash(entry.Finfo.Name())
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
	req := mcstore.DirectoryRequest{
		ProjectName: u.project.Name,
		ProjectID:   u.project.ProjectID,
		Path:        dirPath,
	}
	dirID := u.getDirectoryWithRetry(req)
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
func (u *uploader) getDirectoryWithRetry(req mcstore.DirectoryRequest) string {
	var dirID string
	fn := func() bool {
		var err error
		if dirID, err = u.serverAPI.GetDirectory(req); err != nil {
			return false
		}
		return true
	}
	u.withRetry(fn)
	return dirID
}

// handleFileEntry will handle processing file entries. It will check if the file
// has already been uploaded. If it hasn't it will upload the file to the server.
func (u *uploader) handleFileEntry(entry files.TreeEntry) {
	if dir := u.getDirByPath(filepath.Dir(entry.Path)); dir == nil {
		app.Log.Panicf("Should have found dir")
	} else {
		file := u.getFileByName(entry.Finfo.Name(), dir.ID)
		switch {
		case file == nil:
			u.uploadFile(entry, file, dir)
		case entry.Finfo.ModTime().Unix() > file.LastUpload.Unix():
			u.uploadFile(entry, file, dir)
		default:
			// nothing to do
		}
	}
}

// getDirByPath looks up a path in the local database. It normalizes
// the file separator.
func (u *uploader) getDirByPath(path string) *Directory {
	path = filepath.ToSlash(path)
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
	// TODO: do something with the starting block (its ignored for now)
	var n int
	var err error
	chunkNumber := 1

	f, _ := os.Open(entry.Path)
	defer f.Close()
	buf := make([]byte, 1024*1024)
	totalChunks := numChunks(entry.Finfo.Size())
	var uploadResp *mcstore.UploadChunkResponse
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
			uploadResp = u.sendFlowDataWithRetry(req)
			fmt.Printf("uploadResp = %#v\n", uploadResp)
			if uploadResp.Done {
				break
			}
			chunkNumber++
		}
		if err != nil {
			break
		}
	}

	if err != nil && err != io.EOF {
		app.Log.Errorf("Unable to complete read on file for upload: %s", entry.Path)
	} else {
		// done, so update the database with the entry.
		if file == nil || file.Checksum != checksum {
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
			u.db.UpdateFile(file)
		}
	}
}

// getUploadResponse sends an upload request to the server and gets the response.
func (u *uploader) getUploadResponse(directoryID string, entry files.TreeEntry) (*mcstore.CreateUploadResponse, string) {
	// retry forever
	checksum, _ := file.HashStr(md5.New(), entry.Path)
	chunkSize := int32(1000 * 1000)
	uploadReq := mcstore.CreateUploadRequest{
		ProjectID:   u.project.ProjectID,
		DirectoryID: directoryID,
		FileName:    entry.Finfo.Name(),
		FileSize:    entry.Finfo.Size(),
		ChunkSize:   chunkSize,
		FileMTime:   entry.Finfo.ModTime().Format(time.RFC1123),
		Checksum:    checksum,
	}
	resp := u.createUploadRequestWithRetry(uploadReq)
	return resp, checksum
}

// createUplaodRequestWithRetry will make the server API CreateUploadRequest call. If it fails
// it will retry (dependent on retry settings).
func (u *uploader) createUploadRequestWithRetry(uploadReq mcstore.CreateUploadRequest) *mcstore.CreateUploadResponse {
	var resp *mcstore.CreateUploadResponse
	fn := func() bool {
		var err error
		if resp, err = u.serverAPI.CreateUploadRequest(uploadReq); err != nil {
			fmt.Println("CreateUploadRequest returned err", err)
			return false
		}
		return true
	}
	u.withRetry(fn)
	return resp
}

// numChunks determines the number of chunks that will be sent to the server.
func numChunks(size int64) int32 {
	d := float64(size) / float64(1024*1024)
	n := int(math.Ceil(d))
	return int32(n)
}

// createUplaodRequestWithRetry will make the server API SendFlowData call. If it fails
// it will retry (dependent on retry settings).
func (u *uploader) sendFlowDataWithRetry(req *flow.Request) *mcstore.UploadChunkResponse {
	var resp *mcstore.UploadChunkResponse
	fn := func() bool {
		var err error
		if resp, err = u.serverAPI.SendFlowData(req); err != nil {
			fmt.Println("SendFlowData returned err", err)
			return false
		}
		return true
	}
	u.withRetry(fn)
	return resp
}

// withRetry will retry the given func. The func should return true
// when it is successful. It will retry retryCount times (forever
// if retryCount is set to retryForever). When a call fails it will
// sleep a random amount of time bounded by minWait and maxWait before
// it retries the request.
func (u *uploader) withRetry(fn func() bool) {
	retryCounter := 0
	for {
		if fn() {
			break
		}

		if u.retryCount != retryForever {
			retryCounter++
			if retryCounter > u.retryCount {
				app.Log.Panicf("Retries exceeded aborting")
			}
		}
		u.sleepRandom()
	}
}

func (u *uploader) sleepRandom() {
	// sleep a random amount between minWait and maxWait
	rand.Seed(time.Now().Unix())
	randomSleepTime := rand.Intn(u.maxWait) + u.minWait
	time.Sleep(time.Duration(randomSleepTime))
}
