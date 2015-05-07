package processor

import (
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/quirkey/magick"
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

	m, err := magick.NewFromFile(filePath)
	if err != nil {
		app.Log.Errorf("Image conversion for %s failed: %s", filePath, err)
		return err
	}

	if err := m.ToFile(conversionFile); err != nil {
		app.Log.Errorf("Image conversion for %s failed: %s", filePath, err)
		return err
	}

	return nil
}
