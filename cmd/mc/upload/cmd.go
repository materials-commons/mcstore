package upload

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstored/service/rest/upload"
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
			Value: 5,
			Usage: "Number of simultaneous uploads to perform, defaults to 5",
		},
	},
	Action: Cmd,
}

const oneMeg = 1024 * 1024
const twoMeg = oneMeg * 2
const largeFileSize = oneMeg * 25

var project string

//var pbPool = &pb.Pool{}

// Cmd implements the cli command upload.
func Cmd(c *cli.Context) {
	fmt.Println("upload: ", c.Args())
	if len(c.Args()) != 1 {
		fmt.Println("You must give the directory to walk")
		os.Exit(1)
	}
	dir := c.Args()[0]
	project = c.String("project")
	numThreads := c.Int("parallel")

	_, errc := files.PWalk(dir, numThreads, processFiles)
	if err := <-errc; err != nil {
		fmt.Println("Got error: ", err)
	}
}

func processFiles(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
	for entry := range entries {
		select {
		case result <- sendFile(entry):
		case <-done:
			fmt.Println("Received done stopping...")
		}
	}
}

func sendFile(fileEntry files.TreeEntry) string {
	createUploadRequest()
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
	return ""
}

func createUploadRequest() {
	req := upload.CreateRequest{
		ProjectID:   "9ead5bbf-f7eb-4010-bc1f-e4a063f56226",
		DirectoryID: "c54a77d6-cd6d-4cd1-8f19-44facc761da6",
		FileName:    "abc.txt",
		FileSize:    10,
		FileMTime:   "Thu, 30 Apr 2015 13:10:04 EST",
	}

	var resp upload.CreateResponse
	c := ezhttp.NewClient()
	s, err := c.JSON(&req).JSONPost("http://localhost:5013/upload?apikey=472abe203cd411e3a280ac162d80f1bf", &resp)
	if err != nil {
		fmt.Println("err =", err)
	}
	fmt.Println("s =", s)
	fmt.Printf("%#v\n", resp)
}

func sendFlowChunk(buf []byte) {

}
