package mc

import (
	"fmt"
	osuser "os/user"
	"path/filepath"

	"github.com/materials-commons/config"
)

var u *osuser.User

func init() {
	var err error
	if u, err = osuser.Current(); err != nil {
		panic(fmt.Sprintf("Couldn't determine current user: %s", err))
	}
}

// Home returns the users home directory.
func Home() string {
	return u.HomeDir
}

// APIKey returns the users APIKey.
func APIKey() string {
	val, err := config.GetStringErr("apikey")
	if err != nil {
		panic(fmt.Sprintf("Unable to get APIKey: %s", err))
	}
	return val
}

// ConfigDir returns the path to the configuration directory. This
// directory is located at $HOME/.materialscommons.
func ConfigDir() string {
	return filepath.Join(u.HomeDir, ".materialscommons")
}

// ConfigFile returns the path to the configuration file. This
// path is ConfigDir()/config.json.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// ProjectsFile returns the path to the projects file. This file
// contains the list of projects, their location and id. The file
// is located at ConfigDir()/projects.json.
func ProjectsFile() string {
	return filepath.Join(ConfigDir(), "projects.json")
}
