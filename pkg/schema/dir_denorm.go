package schema

import (
	"time"
)

// FileEntry is a denormalized instance of a datafile used in the datadirs_denorm table.
type FileEntry struct {
	ID        string                 `gorethink:"id"`
	Name      string                 `gorethink:"name"`
	Owner     string                 `gorethink:"owner"`
	Birthtime time.Time              `gorethink:"birthtime"`
	Checksum  string                 `gorethink:"checksum"`
	Size      int64                  `gorethink:"size"`
	Tags      map[string]interface{} `gorethink:"tags"`
	MediaType string                 `gorethink:"mediatype"`
}

// DataDirDenorm is a denormalized instance of a datadir used in the datadirs_denorm table.
type DataDirDenorm struct {
	ID        string                 `gorethink:"id"`
	Name      string                 `gorethink:"name"`
	Owner     string                 `gorethink:"owner"`
	Birthtime time.Time              `gorethink:"birthtime"`
	DataFiles []FileEntry            `gorethink:"datafiles"`
	ProjectID string                 `gorethink:"project_id"`
	Tags      map[string]interface{} `gorethink:"tags"`
}

// Filter will filter out non matching FileEntry items.
func (d DataDirDenorm) Filter(keep func(f FileEntry) bool) []FileEntry {
	var keptEntries []FileEntry
	for _, fileEntry := range d.DataFiles {
		if keep(fileEntry) {
			keptEntries = append(keptEntries, fileEntry)
		}
	}

	return keptEntries
}
