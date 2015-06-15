package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/willf/bitset"
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
	upload.File.Blocks = toBitSet(upload.File.BitString)
	return &upload, nil
}

// Search attempts to find a matching upload request matching the given
// parameters.
func (u rUploads) Search(params UploadSearch) (*schema.Upload, error) {
	rql := model.Uploads.T().GetAllByIndex("project_id", params.ProjectID).
		Filter(r.Row.Field("directory_id").Eq(params.DirectoryID))
	var uploads []schema.Upload
	if err := model.Uploads.Qs(u.session).Rows(rql, &uploads); err != nil {
		return nil, err
	}

	match := func(uitem schema.Upload) bool {
		if uitem.File.Name == params.FileName && uitem.File.Checksum == params.Checksum {
			return true
		}
		return false
	}

	if matchingUpload := schema.Uploads.Find(uploads, match); matchingUpload != nil {
		// Unmarshal into bitset ourselves since it doesn't persist properly to
		// the database.
		matchingUpload.File.Blocks = toBitSet(matchingUpload.File.BitString)
		return matchingUpload, nil
	}

	return nil, app.ErrNotFound
}

// Insert adds a new upload to the uploads table.
func (u rUploads) Insert(upload *schema.Upload) (*schema.Upload, error) {
	var newUpload schema.Upload
	upload.File.BitString = toBitStr(upload.File.Blocks)
	if err := model.Uploads.Qs(u.session).Insert(upload, &newUpload); err != nil {
		return nil, err
	}
	// Since bitset doesn't persist properly, set it ourselves.
	newUpload.File.Blocks = toBitSet(newUpload.File.BitString)
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
	for _, upload := range uploads {
		upload.File.Blocks = toBitSet(upload.File.BitString)
	}
	return uploads, nil
}

// ForProject returns all uploads for a particular.
func (u rUploads) ForProject(projectID string) ([]schema.Upload, error) {
	rql := model.Uploads.T().GetAllByIndex("project_id", projectID)
	var uploads []schema.Upload
	if err := model.Uploads.Qs(u.session).Rows(rql, &uploads); err != nil {
		return nil, err
	}
	for _, upload := range uploads {
		upload.File.Blocks = toBitSet(upload.File.BitString)
	}
	return uploads, nil
}

// Delete deletes the given upload id
func (u rUploads) Delete(uploadID string) error {
	return model.Uploads.Qs(u.session).Delete(uploadID)
}

// toBitSet will turn a bytes representation of bitset back into a BitSet.
func toBitSet(bitstr []byte) *bitset.BitSet {
	b := &bitset.BitSet{}
	b.UnmarshalJSON(bitstr)
	return b
}

// toBitStr will turn a BitSet into a []byte array representation.
func toBitStr(b *bitset.BitSet) []byte {
	bytes, _ := b.MarshalJSON()
	return bytes
}
