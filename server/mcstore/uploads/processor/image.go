package processor

import (
	"os"
	"path/filepath"

	"os/exec"

	"github.com/materials-commons/mcstore/pkg/app"
)

// imageFileProcessor processes image files. It converts
// bmp and tiff files into jpg files so they can be
// displayed on the web.
type imageFileProcessor struct {
	fileID string
}

// newImageFileProcessor creates a processor for converting image types to jpg.
func newImageFileProcessor(fileID string) *imageFileProcessor {
	return &imageFileProcessor{
		fileID: fileID,
	}
}

// process will convert an image to the JPEG format. It stores the image in a
// subdirectory called .conversion located in the directory of the original file.
func (i *imageFileProcessor) Process() error {
	filePath := app.MCDir.FilePath(i.fileID)
	fileDir := app.MCDir.FileDir(i.fileID)
	conversionDir := filepath.Join(fileDir, ".conversion")
	conversionFile := filepath.Join(conversionDir, i.fileID+".jpg")

	if err := os.MkdirAll(conversionDir, 0777); err != nil {
		app.Log.Errorf("Image conversion couldn't create .conversion directory: %s", err)
		return err
	}

	return convert(filePath, conversionFile)
}

func convert(file, conversionFile string) error {
	var (
		err error
		out []byte
	)

	cmd := "convert"
	args := []string{file, conversionFile}
	if out, err = exec.Command(cmd, args...).Output(); err != nil {
		app.Log.Errorf("convert command failed: %s", err)
	}

	app.Log.Debugf("convert command output: %s", string(out))
	return err
}
