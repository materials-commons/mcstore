package mc

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/mcstore/pkg/app"
)

type Project struct {
	ID           int64
	Name         string
	Path         string
	ProjectID    string
	LastUpload   time.Time
	LastDownload time.Time
}

type Directory struct {
	ID           int64
	DirectoryID  string
	Path         string
	LastUpload   time.Time
	LastDownload time.Time
}

type File struct {
	ID           int64
	FileID       string
	Name         string
	Checksum     string
	Size         int64
	MTime        time.Time
	CTime        time.Time
	LastUpload   time.Time
	LastDownload time.Time
	Directory    int64
}

type schemaIndex struct {
	description string
	sql         string
}

type schemaCommand struct {
	description string
	sql         string
	indices     []schemaIndex
}

var schemas = []schemaCommand{
	{
		description: "Create the project table",
		sql: `
                     create table project(
                        id integer primary key,
                        name text,
                        projectid varchar(40),
                        path text,
                        lastupload datetime,
                        lastdownload datetime
                     )
                `,
		indices: []schemaIndex{},
	},
	{
		description: "Create the directories table that holds all known directories",
		sql: `
                      create table directories(
                         id integer primary key,
                         directoryid varchar(40),
                         path text,
                         lastupload datetime,
                         lastdownload datetime
                      )
                `,
		indices: []schemaIndex{
			{
				description: "Create an index for directory paths",
				sql:         "create index directories_path on directories(path)",
			},
			{
				description: "Create an index for the materials commons directory id",
				sql:         "create index directories_dirid on directories(directoryid)",
			},
		},
	},

	{
		description: "Create the table that holds all known files",
		sql: `
             create table files(
                id integer primary key,
                fileid varchar(40),
                name text,
                checksum varchar(40),
                size bigint,
                mtime datetime,
                ctime datetime,
                lastupload datetime,
                lastdownload datetime,
                directory integer,
                foreign key (directory) references directories(id)
            )
        `,
		indices: []schemaIndex{
			{
				description: "Create an index on directory",
				sql:         "create index files_directory on files(directory)",
			},
			{
				description: "Create an index on name",
				sql:         "create index files_name on files(name)",
			},
			{
				description: "Create an index on checksum",
				sql:         "create index files_checksum on files(checksum)",
			},
			{
				description: "Create an index on materials commands file id ",
				sql:         "create index files_fileid on files(fileid)",
			},
		},
	},
}

func createSchema(db *sqlx.DB) error {
	for _, entry := range schemas {
		_, err := db.Exec(entry.sql)
		if err != nil {
			app.Log.Errorf("failed on create for %s/%s: %s", entry.description, entry.sql, err)
			return app.ErrInvalid
		}
		for _, index := range entry.indices {
			_, err := db.Exec(index.sql)
			if err != nil {
				app.Log.Errorf("failed on create for %s/%s: %s", index.description, index.sql, err)
				return app.ErrInvalid
			}
		}
	}
	return nil
}
