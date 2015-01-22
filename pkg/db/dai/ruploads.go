package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// rUploads implements the Uploads interface for RethinkDB.
type rUploads struct {
	session *r.Session
}

// NewRUploads create a new instance of rUploads.
func NewRUploads(session *r.Session) rUploads {
	return rUploads{
		session: session,
	}
}

// ByID looks up an upload by its primary key (id).
func (u rUploads) ByID(id string) (*schema.Upload, error) {
	var upload schema.Upload
	if err := model.Uploads.Qs(u.session).ByID(id, &upload); err != nil {
		return nil, err
	}
	return &upload, nil
}

// Insert adds a new upload to the uploads table.
func (u rUploads) Insert(upload *schema.Upload) (*schema.Upload, error) {
	var newUpload schema.Upload
	if err := model.Uploads.Qs(u.session).Insert(upload, &newUpload); err != nil {
		return nil, err
	}
	return &newUpload, nil
}

// Update updates an existing upload entry.
func (u rUploads) Update(upload *schema.Upload) error {
	if err := model.Uploads.Qs(u.session).Update(upload.ID, upload); err != nil {
		return err
	}
	return nil
}

// ForOwner retrieves all the uploads for the named user.
func (u rUploads) ForUser(user string) ([]schema.Upload, error) {
	rql := model.Uploads.T().GetAllByIndex("owner", user)
	var uploads []schema.Upload
	if err := model.Uploads.Qs(u.session).Rows(rql, &uploads); err != nil {
		return nil, err
	}
	return uploads, nil
}

// Delete deletes the given upload id
func (u rUploads) Delete(uploadID string) error {
	return model.Uploads.Qs(u.session).Delete(uploadID)
}
