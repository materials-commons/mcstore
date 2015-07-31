package search

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
	"gopkg.in/olivere/elastic.v2"
)

func NewProcessesIndexer(client *elastic.Client, session *r.Session) *Indexer {
	getProcessesSetup := func(row r.Term) interface{} {
		return map[string]interface{}{
			"setup": r.Table("process2setup").GetAllByIndex("process_id", row.Field("process_id")).
				EqJoin("setup_id", r.Table("setupproperties"), r.EqJoinOpts{Index: "setup_id"}).
				Zip().CoerceTo("ARRAY"),
		}
	}
	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2process"), r.EqJoinOpts{Index: "project_id"}).
		Zip().
		EqJoin("process_id", r.Table("processes")).Zip().
		Merge(getProcessesSetup)

	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			s := item.(*doc.Process)
			return s.ProcessID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}

func NewSamplesIndexer(client *elastic.Client, session *r.Session) *Indexer {
	propertiesAndFiles := func(row r.Term) interface{} {
		return map[string]interface{}{
			"properties": r.Table("sample2propertyset").
				GetAllByIndex("sample_id", row.Field("sample_id")).
				EqJoin("property_set_id", r.Table("propertyset2property"), r.EqJoinOpts{Index: "property_set_id"}).
				Zip().
				EqJoin("property_id", r.Table("properties")).Zip().Pluck("attribute", "name").
				CoerceTo("ARRAY"),
			"files": r.Table("sample2datafile").GetAllByIndex("sample_id", row.Field("sample_id")).
				EqJoin("datafile_id", r.Table("datafiles")).Zip().CoerceTo("ARRAY"),
		}
	}
	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2sample"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("sample_id", r.Table("samples")).Zip().
		Merge(propertiesAndFiles)

	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			s := item.(*doc.Sample)
			return s.SampleID
		},
		Apply: func(item interface{}) {
			s := item.(*doc.Sample)
			for i, _ := range s.Files {
				f := s.Files[i]
				s.Files[i].Contents = ReadFileContents(f.DataFileID, f.MediaType.Mime, f.Name, f.Size)
			}
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}

func NewProjectsIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("projects")
	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			project := item.(*schema.Project)
			return project.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}

func NewUsersIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("users")

	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			user := item.(*schema.User)
			return user.ID
		},
		Client:   client,
		Session:  session,
		MaxCount: 1000,
	}
}

func NewFilesIndexer(client *elastic.Client, session *r.Session) *Indexer {
	renameDirPath := func(row r.Term) interface{} {
		return row.Merge(map[string]interface{}{
			"right": map[string]interface{}{
				"path": row.Field("right").Field("name"),
			},
		})
	}

	tagsAndNotes := func(row r.Term) interface{} {
		return map[string]interface{}{
			"tags": r.Table("tag2item").GetAllByIndex("item_id", row.Field("id")).
				Pluck("tag_id").CoerceTo("ARRAY"),
			"notes": r.Table("note2item").GetAllByIndex("item_id", row.Field("id")).
				EqJoin("note_id", r.Table("notes")).Zip().CoerceTo("ARRAY"),
		}
	}

	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2datafile"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("datafile_id", r.Table("datadir2datafile"), r.EqJoinOpts{Index: "datafile_id"}).Zip().
		EqJoin("datadir_id", r.Table("datadirs")).
		Map(renameDirPath).
		Zip().
		EqJoin("datafile_id", r.Table("datafiles")).Zip().
		//Filter(r.Row.Field("id").Eq("184e5b21-b86a-4fd0-97ea-98c726a9787b")).
		//Filter(r.Row.Field("id").Eq("b20cde2d-350b-4bc4-8700-e42352bb70df")).
		Merge(tagsAndNotes)

	return &Indexer{
		RQL: rql,
		GetID: func(item interface{}) string {
			dfile := item.(*doc.File)
			return dfile.ID
		},
		Apply: func(item interface{}) {
			dfile := item.(*doc.File)
			dfile.Contents = ReadFileContents(dfile.ID, dfile.MediaType.Mime, dfile.Name, dfile.Size)
		},
		Client:   client,
		Session:  session,
		MaxCount: 10,
	}
}
