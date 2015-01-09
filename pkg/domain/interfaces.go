package domain

import "github.com/materials-commons/mcfs/base/schema"

type Validator interface {
	HasAccess(userID, fileID string) (schema.File, error)
}
