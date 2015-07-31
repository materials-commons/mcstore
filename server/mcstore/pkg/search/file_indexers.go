package search

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
	"gopkg.in/olivere/elastic.v2"
)

func fileTagsAndNotes(row r.Term) interface{} {
	return map[string]interface{}{
		"tags": r.Table("tag2item").GetAllByIndex("item_id", row.Field("id")).
			Pluck("tag_id").CoerceTo("ARRAY"),
		"notes": r.Table("note2item").GetAllByIndex("item_id", row.Field("id")).
			EqJoin("note_id", r.Table("notes")).Zip().CoerceTo("ARRAY"),
	}
}

func fileRenameDirPath(row r.Term) interface{} {
	return row.Merge(map[string]interface{}{
		"right": map[string]interface{}{
			"path": row.Field("right").Field("name"),
		},
	})
}

func NewFilesIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2datafile"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("datafile_id", r.Table("datadir2datafile"), r.EqJoinOpts{Index: "datafile_id"}).Zip().
		EqJoin("datadir_id", r.Table("datadirs")).
		Map(fileRenameDirPath).
		Zip().
		EqJoin("datafile_id", r.Table("datafiles")).Zip().
		//Filter(r.Row.Field("id").Eq("184e5b21-b86a-4fd0-97ea-98c726a9787b")).
		//Filter(r.Row.Field("id").Eq("b20cde2d-350b-4bc4-8700-e42352bb70df")).
		Merge(fileTagsAndNotes)

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

func NewSingleFileIndexer(client *elastic.Client, session *r.Session, fileID string) *Indexer {
	rql := r.Table("project2datafile").GetAllByIndex("datafile_id", fileID).
		EqJoin("datafile_id", r.Table("datadir2datafile"), r.EqJoinOpts{Index: "datafile_id"}).Zip().
		EqJoin("datadir_id", r.Table("datadirs")).
		Map(fileRenameDirPath).
		Zip().
		EqJoin("datafile_id", r.Table("datafiles")).Zip().
		//Filter(r.Row.Field("id").Eq("184e5b21-b86a-4fd0-97ea-98c726a9787b")).
		//Filter(r.Row.Field("id").Eq("b20cde2d-350b-4bc4-8700-e42352bb70df")).
		Merge(fileTagsAndNotes)

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
