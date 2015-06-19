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

type projectUploader struct {
	db         ProjectDB
	numThreads int
}

func (p *projectUploader) upload() error {
	db := p.db
	project := db.Project()

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

type uploader struct {
	db         ProjectDB
	serverAPI  *mcstore.ServerAPI
	project    *Project
	minWait    int
	maxWait    int
	retryCount int
}

const defaultMinWaitBeforeRetry = 100
const defaultMaxWaitBeforeRetry = 5000
const retryForever = -1

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

func (u *uploader) uploadEntries(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
	for entry := range entries {
		select {
		case result <- u.uploadEntry(entry):
		case <-done:
			return
		}
	}
}

func (u *uploader) uploadEntry(entry files.TreeEntry) string {
	switch {
	case entry.Finfo.IsDir():
		u.handleDirEntry(entry)
	default:
		u.handleFileEntry(entry)
	}
	return ""
}

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

func (u *uploader) getUploadResponse(directoryID string, entry files.TreeEntry) (*mcstore.CreateUploadResponse, string) {
	// retry forever
	checksum, _ := file.HashStr(md5.New(), entry.Path)
	chunkSize := int32(1000*1000)
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

func numChunks(size int64) int32 {
	d := float64(size) / float64(1024*1024)
	n := int(math.Ceil(d))
	return int32(n)
}

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
