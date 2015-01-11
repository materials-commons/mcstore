package upload

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

type assembler struct {
	projectID   string
	directoryID string
	fileID      string
	log         log15.Logger
}

func newAssembler(projectID, directoryID, fileID string) assembler {
	return assembler{
		projectID:   projectID,
		directoryID: directoryID,
		fileID:      fileID,
		log:         app.NewLog("go-routine", "assembler"),
	}
}

func newAssemberFromFlowRequest(flowReq *flow.Request) assembler {
	return newAssembler(flowReq.ProjectID, flowReq.DirectoryID, flowReq.FileID)
}

// Sort the chunks by id. Each chunk is named by its chunk number.
type byChunk []os.FileInfo

func (c byChunk) Len() int      { return len(c) }
func (c byChunk) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c byChunk) Less(i, j int) bool {
	chunkIName, _ := strconv.Atoi(c[i].Name())
	chunkJName, _ := strconv.Atoi(c[j].Name())
	return chunkIName < chunkJName
}

// assembleFile reassembles a file by combining all of its chunks. This routine
// needs to ensure that all the chunks have arrived. It does this by counting
// the number of entries in the directory vs the expected number
func (a assembler) assembleFile() {
	uploadPath := fileUploadPath(a.projectID, a.directoryID, a.fileID)
	filePath := app.MCDir.FilePath(a.fileID)
	fdst, err := os.Create(filePath)
	if err != nil {
		a.log.Error(app.Logf("Error creating assembly file %s: %s", filePath, err))
		return
	}
	defer fdst.Close()

	finfos, err := ioutil.ReadDir(uploadPath)
	if err != nil {
		a.log.Error(app.Logf("Error reading assembly dir %s: %s", uploadPath, err))
		return
	}

	sort.Sort(byChunk(finfos))
	for _, finfo := range finfos {
		fsrc, err := os.Open(filepath.Join(uploadPath, finfo.Name()))
		if err != nil {
			a.log.Error(app.Logf("Error reading chunk %s: %s", finfo.Name(), err))
			return
		}
		io.Copy(fdst, fsrc)
		fsrc.Close()
	}
	os.RemoveAll(uploadPath)
}
