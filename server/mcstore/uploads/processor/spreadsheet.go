package processor

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
	return nil
}
