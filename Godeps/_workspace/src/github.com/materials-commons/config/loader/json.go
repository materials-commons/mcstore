package loader

import (
	"bytes"
	"encoding/json"
	"github.com/materials-commons/config/cfg"
	"io"
)

type jsonLoader struct {
	r io.Reader
}

// JSON creates a new Loader for JSON formatted data.
func JSON(r io.Reader) cfg.Loader {
	return &jsonLoader{r: r}
}

// Load loads the data from the reader.
func (l *jsonLoader) Load(out interface{}) error {
	var buf bytes.Buffer
	buf.ReadFrom(l.r)
	if err := json.Unmarshal(buf.Bytes(), out); err != nil {
		return err
	}
	return nil
}
