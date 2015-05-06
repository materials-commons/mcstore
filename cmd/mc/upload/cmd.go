package upload

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstored/service/rest/upload"
	"github.com/parnurzeal/gorequest"
)

// Command contains the arguments and functions for the cli upload command.
var Command = cli.Command{
	Name:    "upload",
	Aliases: []string{"up", "u"},
	Usage:   "Upload data to MaterialsCommons",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "project, p, proj",
			Usage: "Project id to upload to",
		},
		cli.IntFlag{
			Name:  "parallel, n",
			Value: 3,
			Usage: "Number of simultaneous uploads to perform, defaults to 3",
		},
	},
	Action: uploadCLI,
}

const oneMeg = 1024 * 1024
const twoMeg = oneMeg * 2
const largeFileSize = oneMeg * 25
const maxSimultaneous = 5

var project string

//var pbPool = &pb.Pool{}

// uploadCLI implements the cli command upload.
func uploadCLI(c *cli.Context) {
	fmt.Println("upload: ", c.Args())
	if len(c.Args()) != 1 {
		fmt.Println("You must give the directory to walk")
		os.Exit(1)
	}
	dir := c.Args()[0]
	project = c.String("project")
	numThreads := getNumThreads(c)

	_, errc := files.PWalk(dir, numThreads, processFiles)
	if err := <-errc; err != nil {
		fmt.Println("Got error: ", err)
	}
}

// getNumThreads ensures that the number of parallel downloads is valid.
func getNumThreads(c *cli.Context) int {
	numThreads := c.Int("parallel")

	if numThreads < 1 {
		fmt.Println("Simultaneous downloads must be positive: ", numThreads)
		os.Exit(1)
	} else if numThreads > maxSimultaneous {
		fmt.Printf("You may not set simultaneous downloads greater than %d: %d\n", maxSimultaneous, numThreads)
		os.Exit(1)
	}

	return numThreads
}

// processFiles is the callback passed into PWalk. It processes each file, determines
// if it should be uploaded, and if so uploads the file. There can be a maxSimultaneous
// processFiles routines running.
func processFiles(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
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

func (u *uploader) sendFile(fileEntry files.TreeEntry) string {
	u.createUploadRequest()
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

func (u *uploader) createUploadRequest() {
	req := upload.CreateRequest{
		ProjectID:   "9ead5bbf-f7eb-4010-bc1f-e4a063f56226",
		DirectoryID: "c54a77d6-cd6d-4cd1-8f19-44facc761da6",
		FileName:    "abc.txt",
		FileSize:    10,
		FileMTime:   "Thu, 30 Apr 2015 13:10:04 EST",
	}

	var resp upload.CreateResponse
	r, body, errs := u.client.Post(app.MCApi.APIUrl("/upload")).Send(&req).End()
	if err := app.MCApi.APIError(r, errs); err != nil {
		return
	}
	app.MCApi.ToJSON(body, &resp)
	fmt.Printf("%#v\n", resp)
}

func sendFlowChunk(buf []byte) {

}
