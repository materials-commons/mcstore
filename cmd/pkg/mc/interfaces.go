package mc

type ProjectDB interface {
	Project() *Project
	InsertDirectory(dir *Directory) (*Directory, error)
	Directories() []Directory
	Ls(dir Directory) []File
	InsertFile(f *File) (*File, error)
	FindDirectory(path string) (*Directory, error)
	Clone() ProjectDB
}

type ProjectDBSpec struct {
	Name      string
	ProjectID string
	Path      string
}

type ProjectOpenFlags int

const (
	ProjectDBCreate ProjectOpenFlags = iota
	ProjectDBMustExist
)

type ProjectDBOpener interface {
	OpenProjectDB(dbSpec ProjectDBSpec, flags ProjectOpenFlags) (ProjectDB, error)
}

type ProjectDBLister interface {
	// All returns a list of the known ProjectDBs. The ProjectDBs
	// are open.
	All() ([]ProjectDB, error)

	// Create will create a new local project and populate
	// the default database entries. The returned ProjectDB
	// has already been opened.
	Create(project *Project) (ProjectDB, error)
}

type Configer interface {
	APIKey() string
	ConfigDir() string
	ConfigFile() string
}
