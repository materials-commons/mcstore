package mccli

import (
	"crypto/tls"
	"fmt"
	"os"

	"strings"

	"path/filepath"

	"io"

	"time"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/gohandy/ezhttp"
	"github.com/materials-commons/mcstore/cmd/pkg/mc"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/server/mcstore"
	"github.com/parnurzeal/gorequest"
)

// createCommandArgs holds values that won't change and are
// needed during the upload process.
type projectCreateCommandArgs struct {
	projectName   string
	projectID     string
	directoryPath string
	n             int
}

var (
	// Command contains the arguments and function for the cli project create command.
	projectCreateCommand = cli.Command{
		Name:    "create",
		Aliases: []string{"cr", "c"},
		Usage:   "Create a new project",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "directory, dir, d",
				Usage: "The base directory for the project",
			},
			cli.IntFlag{
				Name:  "parallel, n",
				Value: 3,
				Usage: "Number of simultaneous uploads to perform, defaults to 3",
			},
		},
		Action: projectCreateCLI,
	}

	// args contains global values, including arguments from the cli
	// that are needed to create and upload a project.
	args projectCreateCommandArgs
)

// createCLI implements the project create command.
func projectCreateCLI(c *cli.Context) {
	if err := validateArgs(c); err != nil {
		fmt.Println("Invalid arguments:", err)
		os.Exit(1)
	}

	p := mc.ProjectDBSpec{
		Name:      args.projectName,
		Path:      args.directoryPath,
		ProjectID: args.projectID,
	}
	proj, err := mc.ProjectOpener.CreateProjectDB(p)
	if err != nil {
		fmt.Println("Unable to create project:", err)
		os.Exit(1)
	}

	fmt.Println("Indexing project...")
	indexProject(args.directoryPath, proj)
	fmt.Println("Done.")
}

// validate will validate the command line arguments. It will print a message
// and exit if there is a bad argument.
func validateArgs(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return fmt.Errorf("You must supply a name for the project.")
	}
	args.projectName = c.Args()[0]

	if err := createProject(args.projectName); err != nil {
		return err
	}

	args.directoryPath = c.String("directory")
	if err := validateDirectoryPath(args.directoryPath); err != nil {
		return err
	}

	args.n = getNumThreads(c)

	return nil
}

// createProject creates the new project for the user.
func createProject(projectName string) error {
	req := mcstore.CreateProjectRequest{
		Name:         projectName,
		MustNotExist: true,
	}

	var resp mcstore.CreateProjectResponse
	client := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	if err := sendRequest(client, "/projects", req, &resp); err != nil {
		fmt.Println("Unable to create project:", err)
		return err
	}

	args.projectID = resp.ProjectID
	return nil
}

// validateDirectoryPath checks that the given directory path exists.
func validateDirectoryPath(path string) error {
	if path == "" {
		return fmt.Errorf("You must specify a local directory path where the project files are located.")
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("Invalid directory: %s", path)
	}
	return nil
}

// indexProject walks the directory tree and indexes each of the files found. Indexing
// can be performed in parallel.
func indexProject(path string, proj mc.ProjectDB) error {
	fn := func(done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
		indexEntries(proj, done, entries, result)
	}
	walker := &files.PWalker{
		NumParallel: args.n,
		ProcessFn:   fn,
		ProcessDirs: true,
	}
	walker.PWalk(args.directoryPath)
	return nil
}

// indexer holds state information for the different go routines used when
// indexing a project. It also caches directories so that the database
// isn't being hammered looking for directory entries.
type indexer struct {
	client   *gorequest.SuperAgent
	dirs     map[string]*mc.Directory
	proj     mc.ProjectDB
	ezclient *ezhttp.EzClient
}

// indexEntries processes the entries sent along the entries channel. It also
// processes done channel events by exiting the go routine.
func indexEntries(proj mc.ProjectDB, done <-chan struct{}, entries <-chan files.TreeEntry, result chan<- string) {
	i := &indexer{
		client:   gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
		proj:     nil, // proj.Clone(),
		dirs:     make(map[string]*mc.Directory),
		ezclient: ezhttp.NewSSLClient(),
	}
	for entry := range entries {
		select {
		case result <- i.indexEntry(entry):
		case <-done:
			return
		}
	}
}

// indexEntry indexes a tree entry. It handles indexing directories and and files.
func (i *indexer) indexEntry(entry files.TreeEntry) string {
	switch {
	case entry.Finfo.IsDir():
		i.indexDirectory(entry)
	default:
		i.indexFile(entry)
	}
	return ""
}

// indexDirectory indexes a directory. It creates a new directory on the server and saves
// the directory in the database. It also caches the directory entry to be used when indexing
// the files in a directory.
func (i *indexer) indexDirectory(entry files.TreeEntry) error {
	req := mcstore.GetDirectoryRequest{
		Path:      toProjectPath(entry.Path),
		ProjectID: args.projectID,
	}
	var resp mcstore.GetDirectoryResponse
	if err := sendRequest(i.client, "", req, &resp); err != nil {
		return err
	}
	dir := &mc.Directory{
		DirectoryID: resp.DirectoryID,
		Path:        entry.Path,
	}
	dir, _ = i.proj.InsertDirectory(dir)
	i.dirs[entry.Path] = dir
	return nil
}

func (i *indexer) indexFile(entry files.TreeEntry) error {
	dir := i.findFileDirectory(entry)
	uploadRequest, _ := i.createUploadRequest(entry, dir)

	var n int
	var err error
	chunkNumber := 1

	f, _ := os.Open(entry.Path)
	defer f.Close()
	buf := make([]byte, twoMeg)
	for {
		n, err = f.Read(buf)
		if n != 0 {
			// send bytes
			req := &flow.Request{
				FlowChunkNumber:  int32(chunkNumber),
				FlowTotalChunks:  0,
				FlowChunkSize:    int32(n),
				FlowTotalSize:    entry.Finfo.Size(),
				FlowIdentifier:   uploadRequest,
				FlowFileName:     entry.Finfo.Name(),
				FlowRelativePath: "",
				ProjectID:        args.projectID,
				DirectoryID:      dir.DirectoryID,
			}
			params := req.ToParamsMap()
			s, perr := i.ezclient.PostFileBytes(mcstore.Api.Url("/chunk"), entry.Finfo.Name(), "chunkData", buf[:n], params)
			if perr != nil {
				app.Log.Errorf("Posting file chunks failed: %d/%s", s, perr)
			}
		}
		if err != nil {
			break
		}
	}

	if err != nil && err != io.EOF {
		// do something
	}
	return nil
}

func (i *indexer) findFileDirectory(entry files.TreeEntry) *mc.Directory {
	dirpath := filepath.Dir(entry.Path)
	if dir, found := i.dirs[dirpath]; found {
		return dir
	}

	// The directory is not in our cache, so go to database to get it.
	dir, _ := i.proj.FindDirectory(dirpath)
	i.dirs[dirpath] = dir
	return dir
}

func (i *indexer) createUploadRequest(entry files.TreeEntry, dir *mc.Directory) (string, error) {
	req := mcstore.CreateUploadRequest{
		ProjectID:     args.projectID,
		DirectoryID:   dir.DirectoryID,
		DirectoryPath: dir.Path,
		FileName:      entry.Finfo.Name(),
		FileSize:      entry.Finfo.Size(),
		FileMTime:     entry.Finfo.ModTime().Format(time.RFC1123),
	}
	var resp mcstore.CreateUploadResponse
	if err := sendRequest(i.client, "/upload", req, &resp); err != nil {
		return "", err
	}

	return resp.RequestID, nil
}

func toProjectPath(dirpath string) string {
	i := strings.Index(dirpath, args.projectName)
	return dirpath[i:]
}

func sendRequest(client *gorequest.SuperAgent, path string, req interface{}, resp interface{}) error {
	r, body, errs := client.Post(mcstore.Api.Url(path)).Send(req).End()
	if err := mcstore.Api.IsError(r, errs); err != nil {
		fmt.Println("Unable to create project:", err)
		return err
	}

	mcstore.Api.ToJSON(body, resp)
	return nil
}
