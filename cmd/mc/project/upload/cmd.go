package upload

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/cmd/pkg/project"
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

var proj *project.MCProject

//var pbPool = &pb.Pool{}

// uploadCLI implements the cli command upload.
func uploadCLI(c *cli.Context) {
	fmt.Println("upload: ", c.Args())
	if len(c.Args()) != 1 {
		fmt.Println("You must give the directory to walk")
		os.Exit(1)
	}
	dir := c.Args()[0]
	numThreads := getNumThreads(c)

	if !file.IsDir(dir) {
		fmt.Printf("Invalid directory: %s\n", dir)
		os.Exit(1)
	}

	uploadToServer(dir, numThreads)
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

// uploadToServer
func uploadToServer(dir string, numThreads int) {
	var err error
	proj, err = project.Find(dir)
	if err != nil {
		fmt.Println("Unable to locate project dir is in.")
		os.Exit(1)
	}
	fmt.Printf("project = '%s'\n", proj.ID)
	// _, errc := files.PWalk(dir, numThreads, processFiles)
	// if err := <-errc; err != nil {
	// 	fmt.Println("Got error: ", err)
	// }
}

// findDotMCProject will walk up from directory looking for the .mcproject
// directory. If it cannot find it, then the directory isn't in a
// known project. findProject will call os.Exit on any errors or if
// it cannot find a .mcproject directory.
func findDotMCProject(dir string) string {
	// Normalize the directory path, and convert all path separators to a
	// forward slash (/).
	dirPath, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("Bad directory %s: %s", dir, err)
		os.Exit(1)
	}

	dirPath = filepath.ToSlash(dirPath)
	for {
		if dirPath == "/" {
			// Projects at root level not allowed
			fmt.Println("Your directory is not in a project.")
			fmt.Println("Upload a directory in a project or create a project by running the create-project command.")
			os.Exit(1)
		}

		mcprojectDir := filepath.Join(dirPath, ".mcproject")
		if file.IsDir(mcprojectDir) {
			// found it
			return mcprojectDir
		}
		dirPath = filepath.Dir(dirPath)
	}
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

// sendFile needs to:
//   if file is not on server then
//       send file up and send hash up at end
//           -- here we are computing hash as we send blocks up
//   else
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
	fmt.Println("createUploadRequest")
	req := upload.CreateRequest{
		ProjectID:   "9ead5bbf-f7eb-4010-bc1f-e4a063f56226",
		DirectoryID: "c54a77d6-cd6d-4cd1-8f19-44facc761da6",
		FileName:    "abc.txt",
		FileSize:    10,
		FileMTime:   "Thu, 30 Apr 2015 13:10:04 EST",
	}

	var resp upload.CreateResponse
	fmt.Println("url =", app.MCApi.APIUrl("/upload"))
	r, body, errs := u.client.Post(app.MCApi.APIUrl("/upload")).Send(req).End()
	if err := app.MCApi.APIError(r, errs); err != nil {
		fmt.Println("got err from Post:", err)
		return
	}
	app.MCApi.ToJSON(body, &resp)
	fmt.Printf("%#v\n", resp)
}

func sendFlowChunk(buf []byte) {

}
