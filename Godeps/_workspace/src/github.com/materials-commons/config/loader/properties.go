package loader

import (
	p "github.com/dmotylev/goproperties"
	"github.com/materials-commons/config/cfg"
	"io"
)

type propertiesLoader struct {
	r io.Reader
}

// Properties creates a new Loader for properties formatted data.
func Properties(r io.Reader) cfg.Loader {
	return &propertiesLoader{r: r}
}

// Load loads the data from the reader.
func (l *propertiesLoader) Load(out interface{}) error {
	properties := make(p.Properties)
	if err := properties.Load(l.r); err != nil {
		return err
	}
	o := out.(*map[string]string)
	*o = map[string]string(properties)
	return nil
}
