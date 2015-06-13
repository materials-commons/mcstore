package mc

type ProjectDB interface {
	Project() *Project
	InsertDirectory(dir *Directory) (*Directory, error)
	Directories() []Directory
	Ls(dir Directory) []File
	InsertFile(f *File) (*File, error)
	FindDirectory(path string) (*Directory, error)
}

type ProjectDBLister interface {
	// All returns a list of the known ProjectDBs. The ProjectDBs
	// are open.
	All() []ProjectDB
	Create(project *Project) (ProjectDB, error)
}
