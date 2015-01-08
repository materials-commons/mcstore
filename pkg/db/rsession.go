package db

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/config"
)

// RSession creates a new RethinkDB session.
func RSession() (*r.Session, error) {
	return r.Connect(
		r.ConnectOpts{
			Address:  config.GetString("MCDB_CONNECTION"),
			Database: config.GetString("MCDB_NAME"),
		})
}
