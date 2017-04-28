package processor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
)

// officeFileProcessor processes excel spreadsheets. It
// converts the spreadsheet into a csv file.
type officeFileProcessor struct {
	fileID string
}

// newOfficeFileProcessor creates a processor for converting spreadsheets
// to csv files.
func newOfficeFileProcessor(fileID string) *officeFileProcessor {
	return &officeFileProcessor{
		fileID: fileID,
	}
}

// process will convert a spreadsheet to a csv file. It stores the csv file
// in a subdirectory called .conversion located in the directory the original
// spreadsheet file is in.
func (s *officeFileProcessor) Process() error {
	filePath := app.MCDir.FilePath(s.fileID)
	fileDir := app.MCDir.FileDir(s.fileID)
	conversionDir := filepath.Join(fileDir, ".conversion")

	if err := os.MkdirAll(conversionDir, 0777); err != nil {
		app.Log.Errorf("Image conversion couldn't create .conversion directory: %s", err)
		return err
	}

	return s.convert(filePath, conversionDir)
}

func (s *officeFileProcessor) convert(filePath, conversionDir string) error {
	var (
		err        error
		out        []byte
		profileDir string
	)

	tmpDir := os.TempDir()

	if profileDir, err = ioutil.TempDir(tmpDir, "materialscommons"); err != nil {
		app.Log.Errorf("Unable to create temporary dir: %s", err)
		return err
	}

	defer os.RemoveAll(profileDir)

	cmd := "libreoffice"
	args := []string{"-env:UserInstallation=file://" + profileDir, "--headless", "--convert-to", "pdf", "--outdir", tmpDir, filePath}
	if out, err = exec.Command(cmd, args...).Output(); err != nil {
		app.Log.Errorf("convert command failed: %s", err)
		return err
	}

	filename := filepath.Base(filePath + ".pdf")
	err = s.moveToConvDir(tmpDir, filename, conversionDir)

	app.Log.Debugf("convert command output: %s", string(out))
	return err
}

func (s *officeFileProcessor) moveToConvDir(tmpFileDir, filename, conversionDir string) error {
	tmpFilePath := filepath.Join(tmpFileDir, filename)
	convDirFilePath := filepath.Join(conversionDir, filename)
	if err := file.Copy(tmpFilePath, convDirFilePath); err != nil {
		return err
	}

	return os.Remove(tmpFilePath)
}
