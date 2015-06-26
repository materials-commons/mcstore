package mcstore

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

func updateKeyCacheOnChange(session *r.Session, keycache *apikeyCache) {
	var c struct {
		NewUserValue schema.User `gorethink:"new_val"`
		OldUserValue schema.User `gorethink:"old_val"`
	}

	users, _ := r.Table("users").Changes().Run(session)
	for users.Next(&c) {
		switch {
		case c.OldUserValue.ID == "":
			// no old id, so new user added
			keycache.addKey(c.NewUserValue.APIKey, &c.NewUserValue)

		case c.OldUserValue.APIKey != c.NewUserValue.APIKey:
			// APIKey changed - reset entry to new value.
			keycache.resetKey(c.OldUserValue.APIKey, c.NewUserValue.APIKey, &c.NewUserValue)
		}
	}
}