package doc

import "time"

type SetupProperties struct {
	ID          string      `gorethink:"-" json:"-"`
	Attribute   string      `gorethink:"attribute" json:"attribute"`
	Description string      `gorethink:"description" json:"description"`
	Name        string      `gorethink:"name" json:"name"`
	ProcessID   string      `gorethink:"-" json:"-"`
	SetupID     string      `gorethink:"-" json:"-"`
	Units       string      `gorethink:"units" json:"units"`
	Value       interface{} `gorethink:"value" json:"value"`
}

type Process struct {
	ID            string            `gorethink:"id" json:"-"`
	Type          string            `gorethink:"_type" json:"_type"`
	Birthtime     time.Time         `gorethink:"birthtime" json:"birthtime"`
	MTime         time.Time         `gorethink:"mtime" json:"mtime"`
	Name          string            `gorethink:"name" json:"name"`
	DoesTransform bool              `gorethink:"does_transform" json:"does_transform"`
	Owner         string            `gorethink:"owner" json:"owner"`
	ProcessID     string            `gorethink:"process_id" json:"process_id"`
	ProcessType   string            `gorethink:"procss_type" json:"process_type"`
	ProjectID     string            `gorethink:"project_id" json:"project_id"`
	What          string            `gorethink:"what" json:"what"`
	Why           string            `gorethink:"why" json:"why"`
	Setup         []SetupProperties `gorethink:"setup" json:"setup"`
}
