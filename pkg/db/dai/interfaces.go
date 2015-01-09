package dai

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Users interface {
	ByAPIKey(apikey string) (schema.User, error)
}

type Files interface {
	ByID(id string) (schema.File, error)
}

type Groups interface {
	HasAccess(owner, user string) bool
}
