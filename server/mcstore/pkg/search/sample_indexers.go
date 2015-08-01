package search

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
	"gopkg.in/olivere/elastic.v2"
)

func samplePropertiesAndFiles(row r.Term) interface{} {
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

func NewSamplesIndexer(client *elastic.Client, session *r.Session) *Indexer {
	rql := r.Table("projects").Pluck("id").
		EqJoin("id", r.Table("project2sample"), r.EqJoinOpts{Index: "project_id"}).Zip().
		EqJoin("sample_id", r.Table("samples")).Zip().
		Merge(samplePropertiesAndFiles)
	indexer := defaultSampleIndexer(client, session)
	indexer.RQL = rql
	return indexer
}

func NewSingleSampleIndexer(client *elastic.Client, session *r.Session, sampleID string) *Indexer {
	rql := r.Table("project2sample").GetAllByIndex("sample_id", sampleID).
		EqJoin("sample_id", r.Table("samples")).Zip().
		Merge(samplePropertiesAndFiles)
	indexer := defaultSampleIndexer(client, session)
	indexer.RQL = rql
	return indexer
}

func defaultSampleIndexer(client *elastic.Client, session *r.Session) *Indexer {
	return &Indexer{
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
