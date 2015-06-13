package mc

import (
	"fmt"
	osuser "os/user"
	"path/filepath"

	"github.com/materials-commons/config"
)

type user struct {
	*osuser.User
}

var User user

func init() {
	var err error
	if User.User , err = osuser.Current(); err != nil {
		panic(fmt.Sprintf("Couldn't determine current user: %s", err))
	}
}

// Home returns the users home directory.
func (u user) Home() string {
	return u.HomeDir
}

// APIKey returns the users APIKey.
func (u user) APIKey() string {
	val, err := config.GetStringErr("apikey")
	if err != nil {
		panic(fmt.Sprintf("Unable to get APIKey: %s", err))
	}
	return val
}

// ConfigDir returns the path to the configuration directory. This
// directory is located at $HOME/.materialscommons.
func (u user) ConfigDir() string {
	return filepath.Join(u.HomeDir, ".materialscommons")
}

// ConfigFile returns the path to the configuration file. This
// path is ConfigDir()/config.json.
func (u user) ConfigFile() string {
	return filepath.Join(u.ConfigDir(), "config.json")
}

// ProjectsFile returns the path to the projects file. This file
// contains the list of projects, their location and id. The file
// is located at ConfigDir()/projects.json.
func (u user) ProjectsFile() string {
	return filepath.Join(u.ConfigDir(), "projects.json")
}
