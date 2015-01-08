package schema

// Project2DataDir is a join table that maps projects to their datadirs.
type Project2DataDir struct {
	ID        string `gorethink:"id,omitempty" db:"-"`
	ProjectID string `gorethink:"project_id" db:"project_id"`
	DataDirID string `gorethink:"datadir_id" db:"datadir_id"`
}

// Project2DataFile is a join table that maps projects to their files.
type Project2DataFile struct {
	ID         string `gorethink:"id,omitempty"`
	ProjectID  string `gorethink:"project_id"`
	DataFileID string `gorethink:"datafile_id"`
}

// DataDir2DataFile is a join table that maps datadirs to their files.
type DataDir2DataFile struct {
	ID         string `gorethink:"id,omitempty"`
	DataDirID  string `gorethink:"datadir_id"`
	DataFileID string `gorethink:"datafile_id"`
}
