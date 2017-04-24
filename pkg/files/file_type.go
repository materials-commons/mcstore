package files

func IsOfficeDocument(mime string) bool {
	switch {
	case isWordDocument(mime):
		return true
	case isExcelDocument(mime):
		return true
	case isPowerpointDocument(mime):
		return true
	default:
		return false
	}
}

var wordMimeTypes = map[string]bool{
	"application/msword":                                                      true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
}

func isWordDocument(mime string) bool {
	_, ok := wordMimeTypes[mime]
	return ok
}

var excelMimeTypes = map[string]bool{
	"application/vnd.ms-excel":                                          true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

func isExcelDocument(mime string) bool {
	_, ok := excelMimeTypes[mime]
	return ok
}

var pptMimeTypes = map[string]bool{
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
}

func isPowerpointDocument(mime string) bool {
	_, ok := pptMimeTypes[mime]
	return ok
}
