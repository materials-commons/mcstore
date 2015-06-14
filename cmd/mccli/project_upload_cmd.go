package mccli

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstore"
	"github.com/parnurzeal/gorequest"
)

// Command contains the arguments and functions for the cli upload command.
var projectUploadCommand = cli.Command{
	Name:    "upload",
	Aliases: []string{"up", "u"},
	Usage:   "Upload data to MaterialsCommons",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "parallel, n",
			Value: 3,
			Usage: "Number of simultaneous uploads to perform, defaults to 3",
		},
	},
	Action: projectUploadCLI,
}

var proj mc.ProjectDB

//var pbPool = &pb.Pool{}

// uploadCLI implements the cli command upload.
func projectUploadCLI(c *cli.Context) {
	fmt.Println("upload: ", c.Args())
	if len(c.Args()) != 1 {
		fmt.Println("You must give a directory to upload.")
		os.Exit(1)
	}
	dir := c.Args()[0]
	numThreads := getNumThreads(c)

	if !file.IsDir(dir) {
		fmt.Printf("Invalid directory: %s.\n", dir)
		os.Exit(1)
	}

	uploadToServer(dir, numThreads)
}

// uploadToServer
func uploadToServer(dir string, numThreads int) {
	var err error
	proj, err = mc.Find(dir)
	if err != nil {
		fmt.Println("Unable to locate project dir is in.")
		os.Exit(1)
	}
	//fmt.Printf("project = '%s'\n", proj.id)
	// _, errc := files.PWalk(dir, numThreads, processFiles)
	// if err := <-errc; err != nil {
	// 	fmt.Println("Got error: ", err)
	// }
}

// processFiles is the callback passed into PWalk. It processes each file, determines
// if it should be uploaded, and if so uploads the file. There can be a maxSimultaneous
// processFiles routines running.
func processFiles(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
	fmt.Println("processFiles")
	u := &uploader{
		client: gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
	}
	for entry := range entries {
		select {
		case result <- u.sendFile(entry):
		case <-done:
			// Received done, so stop processing requests.
			return
		}
	}
}

type uploader struct {
	client *gorequest.SuperAgent
}

// sendFile logic:
// if file has changed || file not in local database then
//       compute hash
//       if hash is on server then
//          tell server to create a new entry pointing to
//          the already uploaded file
//       else
//          upload the file
//       end
//  end
//
func (u *uploader) sendFile(fileEntry files.TreeEntry) string {
	u.createUploadRequest()
	//s1 := file.ExInfoFrom(20, time.Now(), time.Now(), time.Now(), file.FID{})
	//fileChanged(s1, s1)
	// buf := make([]byte, twoMeg)
	// f, err := os.Open(fileEntry.Path)
	// if err != nil {
	// 	return ""
	// }
	// fileHash := ""
	// // For small files just transfer the bytes. For
	// // large files compute the file hash.
	// if fileEntry.Finfo.Size() > largeFileSize {
	// 	fileHash, _ = file.HashStr(md5.New(), fileEntry.Path)
	// }
	// for {
	// 	read, err := f.Read(buf)
	// 	sendFlowChunk(buf)
	// }
	return fileEntry.Path
}

func fileChanged(oinfo, ninfo file.ExFileInfo) bool {
	switch {
	case oinfo.Size() != ninfo.Size():
		return true
	case oinfo.CTime().Before(ninfo.CTime()):
		return true
	case oinfo.ModTime().Before(ninfo.ModTime()):
		return true
	default:
		return false
	}
}

func (u *uploader) createUploadRequest() {
	fmt.Println("createUploadRequest")
	req := mcstore.CreateUploadRequest{
		ProjectID:   "9ead5bbf-f7eb-4010-bc1f-e4a063f56226",
		DirectoryID: "c54a77d6-cd6d-4cd1-8f19-44facc761da6",
		FileName:    "abc.txt",
		FileSize:    10,
		FileMTime:   "Thu, 30 Apr 2015 13:10:04 EST",
	}

	var resp mcstore.CreateUploadResponse
	fmt.Println("url =", mc.Api.Url("/upload"))
	r, body, errs := u.client.Post(mc.Api.Url("/upload")).Send(req).End()
	if err := mc.Api.IsError(r, errs); err != nil {
		fmt.Println("got err from Post:", err)
		return
	}
	mc.Api.ToJSON(body, &resp)
	fmt.Printf("%#v\n", resp)
}

func sendFlowChunk(buf []byte) {

}
