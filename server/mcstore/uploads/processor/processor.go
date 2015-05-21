package processor

import "github.com/materials-commons/mcstore/pkg/db/schema"

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

	switch mediatype.Mime {
	case "image/tiff":
		return newImageFileProcessor(fileID)
	case "image/bmp":
		return newImageFileProcessor(fileID)
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return newSpreadsheetFileProcessor(fileID)
	case "application/vnd.MS-Excel":
		return newSpreadsheetFileProcessor(fileID)
	default:
		// Not a file type we process (yet)
		return &noopFileProcessor{}
	}
}
