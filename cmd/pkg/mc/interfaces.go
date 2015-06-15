package mc

type ProjectDB interface {
	Project() *Project
	UpdateProject(project *Project) error
	InsertDirectory(dir *Directory) (*Directory, error)
	UpdateDirectory(dir *Directory) error
	Directories() []Directory
	Ls(dir Directory) []File
	InsertFile(f *File) (*File, error)
	UpdateFile(f *File) error
	FindDirectory(path string) (*Directory, error)
	Clone() ProjectDB
}

type ProjectDBSpec struct {
	Name      string
	ProjectID string
	Path      string
}

type ProjectOpenFlags int

type ProjectDBOpener interface {
	CreateProjectDB(dbSpec ProjectDBSpec) (ProjectDB, error)
	OpenProjectDB(name string) (ProjectDB, error)
	PathToName(path string) string
}

type ProjectDBLister interface {
	// All returns a list of the known ProjectDBs. The ProjectDBs
	// are open.
	All() ([]ProjectDB, error)

	// Create will create a new local project and populate
	// the default database entries. The returned ProjectDB
	// has already been opened.
	Create(dbSpec ProjectDBSpec) (ProjectDB, error)
}

type Configer interface {
	APIKey() string
	ConfigDir() string
	ConfigFile() string
}
