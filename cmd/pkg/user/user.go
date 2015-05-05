package user

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

func Home() string {
	return u.HomeDir
}

func APIKey() string {
	val, err := config.GetStringErr("apikey")
	if err != nil {
		panic(fmt.Sprintf("Unable to get APIKey: %s", err))
	}
	return val
}

func ConfigDir() string {
	return filepath.Join(u.HomeDir, ".materialscommons")
}

func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.json")
}
