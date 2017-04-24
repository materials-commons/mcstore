package processor

import (
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/files"
)

// fileProcess defines an interface for processing different
// types of files. Processing may include extracting data,
// conversion of the file to a different type, or whatever
// is deemed appropriate for the file type.
type Processor interface {
	Process() error
}

// newFileProcessor creates a new instance of a fileProcessor. It looks at
// the mime type for the file to determine what kind of processor it should
// use to handle this file. By default it returns a processor that does
// nothing to the file.
func New(fileID string, mediatype schema.MediaType) Processor {
	switch {
	case isImageTypeNeedingConversion(mediatype.Mime):
		return newImageFileProcessor(fileID)
	case files.IsOfficeDocument(mediatype.Mime):
		return newOfficeFileProcessor(fileID)
	default:
		// Not a file type we process (yet)
		return &noopFileProcessor{}
	}
}

func isImageTypeNeedingConversion(mime string) bool {
	switch mime {
	case "image/tiff", "image/bmp":
		return true
	default:
		return false
	}
}
