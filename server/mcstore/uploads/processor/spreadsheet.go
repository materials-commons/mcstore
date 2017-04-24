package processor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
)

// spreadsheetFileProcessor processes excel spreadsheets. It
// converts the spreadsheet into a csv file.
type spreadsheetFileProcessor struct {
	fileID string
}

// newSpreadsheetFileProcessor creates a processor for converting spreadsheets
// to csv files.
func newSpreadsheetFileProcessor(fileID string) *spreadsheetFileProcessor {
	return &spreadsheetFileProcessor{
		fileID: fileID,
	}
}

// process will convert a spreadsheet to a csv file. It stores the csv file
// in a subdirectory called .conversion located in the directory the original
// spreadsheet file is in.
func (s *spreadsheetFileProcessor) Process() error {
	filePath := app.MCDir.FilePath(s.fileID)
	fileDir := app.MCDir.FileDir(s.fileID)
	conversionDir := filepath.Join(fileDir, ".conversion")

	if err := os.MkdirAll(conversionDir, 0777); err != nil {
		app.Log.Errorf("Image conversion couldn't create .conversion directory: %s", err)
		return err
	}

	return s.convert(filePath, conversionDir)
}

func (s *spreadsheetFileProcessor) convert(filePath, conversionDir string) error {
	var (
		err        error
		out        []byte
		profileDir string
	)

	if profileDir, err = ioutil.TempDir(os.TempDir(), "materialscommons"); err != nil {
		app.Log.Errorf("Unable to create temporary dir: %s", err)
		return err
	}

	defer os.RemoveAll(profileDir)

	cmd := "libreoffice"
	args := []string{"-env:UserInstallation=file://" + profileDir, "--headless", "--convert-to", "pdf", "--outdir", conversionDir, filePath}
	if out, err = exec.Command(cmd, args...).Output(); err != nil {
		app.Log.Errorf("convert command failed: %s", err)
		return err
	}

	app.Log.Debugf("convert command output: %s", string(out))

	err = os.Rename(filePath, filepath.Join(conversionDir, filePath+".pdf"))
	return err
}
