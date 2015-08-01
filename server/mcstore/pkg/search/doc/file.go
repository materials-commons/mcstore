package doc

import "github.com/materials-commons/mcstore/pkg/db/schema"

type TagID struct {
	TagID string `gorethink:"tag_id" json:"tag"`
}

type Note struct {
	ID    string `gorethink:"id" json:"id"`
	Note  string `gorethink:"note" json:"note"`
	Title string `gorethink:"title" json:"title"`
}

type File struct {
	schema.File
	Tags      []TagID `gorethink:"tags" json:"tags"`
	DataDirID string  `gorethink:"datadir_id" json:"datadir_id"`
	ProjectID string  `gorethink:"project_id" json:"project_id"`
	Contents  string  `gorethink:"-" json:"contents"` // Contents of the file (text only)
	Notes     []Note  `gorethink:"notes" json:"notes"`
}