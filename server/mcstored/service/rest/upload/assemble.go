package upload

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
)

type assembler struct {
	request *flow.Request
	log     log15.Logger
	tracker *uploadTracker
}

func newAssembler(request *flow.Request, tracker *uploadTracker) *assembler {
	return &assembler{
		request: request,
		tracker: tracker,
		log:     app.NewLog("go-routine", "assembler"),
	}
}

func (a *assembler) launch() {
	go func() {
		if a.assembleFile() {
			a.finish()
		}
	}()
}

// assembleFile reassembles a file by combining all of its chunks. This routine
// needs to ensure that all the chunks have arrived. It does this by counting
// the number of entries in the directory vs the expected number
func (a *assembler) assembleFile() bool {
	if fdst, err := a.createUploadPaths(); err == nil {
		defer fdst.Close()
		return a.assembleFromDirectory(fdst)
	}
	return false
}

func (a *assembler) createUploadPaths() (*os.File, error) {
	os.MkdirAll(app.MCDir.FileDir(a.request.FileID), 0700)
	filePath := app.MCDir.FilePath(a.request.FileID)
	fdst, err := os.Create(filePath)
	if err != nil {
		a.log.Error(app.Logf("Error creating assembly file %s: %s", filePath, err))
	}
	return fdst, err
}

func (a *assembler) assembleFromDirectory(fdst *os.File) bool {
	uploadDir := a.request.Dir()
	finfos, err := ioutil.ReadDir(uploadDir)
	if err != nil {
		a.log.Error(app.Logf("Error reading assembly dir %s: %s", uploadDir, err))
		return false
	}

	//sort.Sort(byChunk(finfos))
	for _, finfo := range finfos {
		fsrc, err := os.Open(filepath.Join(uploadDir, finfo.Name()))
		if err != nil {
			a.log.Error(app.Logf("Error reading chunk %s: %s", finfo.Name(), err))
			return false
		}
		io.Copy(fdst, fsrc)
		fsrc.Close()
	}
	os.RemoveAll(uploadDir)
	return true
}

// finish takes care of the final steps. At this point the file
// has been assembled. There are a couple of other checks that
// need to be done:
//    1. Create the file hash and check if it already exists
//    2. If hash exists, delete file and mark its usesid, mark
//       the upload as finished
//    3. If no hash, then mark the upload as finished.
func (a *assembler) finish() {
	//
	a.tracker.clear(a.request.UploadID())
}
