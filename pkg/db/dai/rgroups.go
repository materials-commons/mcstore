package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// rGroups implements the Groups interface for RethinkDB
type rGroups struct {
	session *r.Session
}

// newRGroups creates a new instance of rGroups
func NewRGroups(session *r.Session) rGroups {
	return rGroups{
		session: session,
	}
}

// ByID looks up a group by its primary key.
func (g rGroups) ByID(id string) (*schema.Group, error) {
	var group schema.Group
	if err := model.Groups.Qs(g.session).ByID(id, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// ForOwner returns all the groups created by this owner
func (g rGroups) ForOwner(owner string) ([]schema.Group, error) {
	rql := model.Groups.T().GetAllByIndex("owner", owner)
	var groups []schema.Group
	if err := model.Groups.Qs(g.session).Rows(rql, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}
