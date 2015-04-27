package model

import (
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

// Groups is a default model for the usergroups table.
var Groups = &rModel{
	schema: schema.Group{},
	table:  "usergroups",
}

// Users is a default model for the users table.
var Users = &rModel{
	schema: schema.User{},
	table:  "users",
}

// Dirs is a default model for the datadirs table.
var Dirs = &rModel{
	schema: schema.Directory{},
	table:  "datadirs",
}

// Files is a default model for the datafiles table
var Files = &rModel{
	schema: schema.File{},
	table:  "datafiles",
}

// Projects is a default model for the projects table
var Projects = &rModel{
	schema: schema.Project{},
	table:  "projects",
}

// Project files
var ProjectFiles = &rModel{
	schema: schema.Project2DataFile{},
	table:  "project2datafile",
}

// Project directories
var ProjectDirs = &rModel{
	schema: schema.Project2DataDir{},
	table:  "project2datadir",
}

// Directory files
var DirFiles = &rModel{
	schema: schema.DataDir2DataFile{},
	table:  "datadir2datafile",
}

// Uploads
var Uploads = &rModel{
	schema: schema.Upload{},
	table:  "uploads",
}

// Access
var Access = &rModel{
	schema: schema.Access{},
	table:  "access",
}
