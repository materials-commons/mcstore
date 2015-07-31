package doc

import (
	"time"

	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type Property struct {
	Attribute string `gorethink:"attribute" json:"attribute"`
	Name      string `gorethink:"name" json:"name"`
}

type SampleFile struct {
	DataFileID string           `gorethink:"datafile_id" json:"datafile_id"`
	Name       string           `gorethink:"name" json:"name"`
	MediaType  schema.MediaType `gorethink:"mediatype" json:"mediatype"`
	Size       int64            `gorethink:"size" json:"size"`
	Contents   string           `gorethink:"-" json:"contents"` // Contents of the file (text only)
}

type Sample struct {
	ID          string       `gorethink:"id" json:"id"`
	Type        string       `gorethink:"_type" json:"_type"`
	Description string       `gorethink:"description" json:"description"`
	Birthtime   time.Time    `gorethink:"birthtime" json:"birthtime"`
	MTime       time.Time    `gorethink:"mtime" json:"mtime"`
	Owner       string       `gorethink:"owner" json:"owner"`
	Name        string       `gorethink:"name" json:"name"`
	ProjectID   string       `gorethink:"project_id" json:"project_id"`
	SampleID    string       `gorethink:"sample_id" json:"sample_id"`
	Properties  []Property   `gorethink:"properties" json:"properties"`
	Files       []SampleFile `gorethink:"files" json:"files"`
}
