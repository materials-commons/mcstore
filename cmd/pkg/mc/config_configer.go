package mc

import (
	"path/filepath"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
)

type configConfiger struct{}

func (c configConfiger) APIKey() string {
	val, err := config.GetStringErr("apikey")
	if err != nil {
		app.Log.Panicf("Unable to get apikey: %s", err)
	}
	return val
}

func (c configConfiger) ConfigDir() string {
	val, err := config.GetStringErr("mcconfigdir")
	if err != nil {
		app.Log.Panicf("Unable to get mcconfigdir: %s", err)
	}
	return val
}

func (c configConfiger) ConfigFile() string {
	return filepath.Join(c.ConfigDir(), "config.json")
}
