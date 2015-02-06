package loader

import (
	"bytes"
	"github.com/materials-commons/config/cfg"
	"gopkg.in/yaml.v1"
	"io"
)

type yamlLoader struct {
	r io.Reader
}

// YAML creates a new Loader for YAML formatted data.
func YAML(r io.Reader) cfg.Loader {
	return &yamlLoader{r: r}
}

// Load loads the data from the reader.
func (l *yamlLoader) Load(out interface{}) error {
	var buf bytes.Buffer
	buf.ReadFrom(l.r)
	if err := yaml.Unmarshal(buf.Bytes(), out); err != nil {
		return err
	}
	return nil
}
