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

// DirsDenorm is a default model for the denormalized datadirs_denorm table
var DirsDenorm = &rModel{
	schema: schema.DataDirDenorm{},
	table:  "datadirs_denorm",
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
