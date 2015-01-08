package schema

// Note represents a note entry.
type Note struct {
	Date    string `gorethink:"date"`
	Message string `gorethink:"message"`
	Who     string `gorethink:"who"`
}
