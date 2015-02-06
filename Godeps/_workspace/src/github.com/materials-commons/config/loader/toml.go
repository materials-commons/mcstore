package loader

import (
	"github.com/BurntSushi/toml"
	"github.com/materials-commons/config/cfg"
	"io"
)

type tomlLoader struct {
	r io.Reader
}

// TOML creates a new Loader for TOML formatted data.
func TOML(r io.Reader) cfg.Loader {
	return &tomlLoader{r: r}
}

// Load loads the data from the reader.
func (l *tomlLoader) Load(out interface{}) error {
	if _, err := toml.DecodeReader(l.r, out); err != nil {
		return err
	}
	return nil
}
