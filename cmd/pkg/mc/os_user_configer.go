package mc

import (
	osuser "os/user"
	"path/filepath"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
)

type osUserConfiger struct {
	*osuser.User
}

func NewOSUserConfiger() osUserConfiger {
	var u osUserConfiger
	if osu, err := osuser.Current(); err != nil {
		app.Log.Panicf("Couldn't determine current user: %s", err)
	} else {
		u.User = osu
	}
	return u
}

// APIKey returns the users APIKey.
func (u osUserConfiger) APIKey() string {
	val, err := config.GetStringErr("apikey")
	if err != nil {
		app.Log.Panicf("Unable to get APIKey: %s", err)
	}
	return val
}

// ConfigDir returns the path to the configuration directory. This
// directory is located at $HOME/.materialscommons.
func (u osUserConfiger) ConfigDir() string {
	return filepath.Join(u.HomeDir, ".materialscommons")
}

// ConfigFile returns the path to the configuration file. This
// path is ConfigDir()/config.json.
func (u osUserConfiger) ConfigFile() string {
	return filepath.Join(u.ConfigDir(), "config.json")
}
