package mc

import (
	"math/rand"
	"time"

	"crypto/md5"
	"io"
	"math"
	"os"

	"github.com/golang/go/src/path/filepath"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstore"
)

type projectUploader struct {
	db         ProjectDB
	numThreads int
}

func (p *projectUploader) upload() error {
	db := p.db
	project := db.Project()

	fn := func(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
		u := &uploader{
			db:        p.db.Clone(),
			serverAPI: mcstore.NewServerAPI(),
			project:   project,
		}
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
	db        ProjectDB
	serverAPI *mcstore.ServerAPI
	project   *Project
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
	// Loop forever asking server for the directory
	dirPath := filepath.ToSlash(entry.Finfo.Name())
	req := mcstore.DirectoryRequest{
		ProjectName: u.project.Name,
		ProjectID:   u.project.ProjectID,
		Path:        dirPath,
	}
	for {
		if dirID, err := u.serverAPI.GetDirectory(req); err != nil {
			// sleep a random amount of time and then retry the request
			sleepRandom()
		} else {
			dir := &Directory{
				DirectoryID: dirID,
				Path:        dirPath,
			}
			if _, err := u.db.InsertDirectory(dir); err != nil {
				app.Log.Panicf("Local database returned err, panic!: %s", err)
			}
			return
		}
	}
}

func (u *uploader) handleFileEntry(entry files.TreeEntry) {
	if dir := u.getDirByPath(filepath.Dir(entry.Path)); dir == nil {
		app.Log.Panicf("Should have found dir")
	} else {
		file := u.getFileByName(entry.Finfo.Name(), dir.ID)
		switch {
		case file == nil:
			u.uploadFile(entry, dir)
		case entry.Finfo.ModTime().Unix() > file.LastUpload.Unix():
			u.uploadFile(entry, dir)
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

func (u *uploader) uploadFile(entry files.TreeEntry, dir *Directory) {
	uploadResponse := u.getUploadResponse(dir.DirectoryID, entry)
	requestID := uploadResponse.RequestID
	// TODO: do something with the starting block (its ignored for now)
	var n int
	var err error
	chunkNumber := 1

	f, _ := os.Open(entry.Path)
	defer f.Close()
	buf := make([]byte, 1024*1024)
	totalChunks := numChunks(entry.Finfo.Size())
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
			u.sendFlowReq(req)
			chunkNumber++
		}
		if err != nil {
			break
		}
	}

	if err != nil && err != io.EOF {
		// do something
	} else {
		// done, so update the database with the entry.
	}
}

func (u *uploader) getUploadResponse(directoryID string, entry files.TreeEntry) *mcstore.CreateUploadResponse {
	// retry forever
	checksum, _ := file.HashStr(md5.New(), entry.Path)
	uploadReq := mcstore.CreateUploadRequest{
		ProjectID:   u.project.ProjectID,
		DirectoryID: directoryID,
		FileName:    entry.Finfo.Name(),
		FileSize:    entry.Finfo.Size(),
		ChunkSize:   1024 * 1024,
		FileMTime:   entry.Finfo.ModTime().Format(time.RFC1123),
		Checksum:    checksum,
	}

	for {
		if resp, err := u.serverAPI.CreateUploadRequest(uploadReq); err != nil {
			sleepRandom()
		} else {
			return resp
		}
	}
	return nil
}

func numChunks(size int64) int32 {
	d := float64(size) / float64(1024*1024)
	n := int(math.Ceil(d))
	return int32(n)
}

func (u *uploader) sendFlowReq(req *flow.Request) {
	// try forever
	for {
		if err := u.serverAPI.SendFlowData(req); err != nil {
			sleepRandom()
		} else {
			break
		}
	}
}

func sleepRandom() {
	// sleep a random amount between 1 and 10 seconds
	rand.Seed(time.Now().Unix())
	randomSleepTime := rand.Intn(10) + 1
	time.Sleep(int64(randomSleepTime))
}
