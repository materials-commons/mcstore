package mcstore

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
	"gopkg.in/olivere/elastic.v2"
)

type process struct {
	ID string `gorethink:"id"`
}

func processChangeIndexer(client *elastic.Client, session *r.Session) {
	var change struct {
		OldProcessValue process `gorethink:"old_val"`
		NewProcessValue process `gorethink:"new_val"`
	}
	var processDoc doc.Process

	processes, _ := r.Table("processes").Changes().Run(session)
	for processes.Next(&change) {
		id := getProcessID(change.OldProcessValue, change.NewProcessValue)
		indexer := search.NewSingleProcessIndexer(client, session, id)
		indexer.Do("processes", processDoc)
	}
}

func getProcessID(oldProcess, newProcess process) string {
	if oldProcess.ID != "" {
		return oldProcess.ID
	}
	return newProcess.ID
}

func fileChangeIndexer(client *elastic.Client, session *r.Session) {
	var change struct {
		OldFileValue schema.File `gorethink:"old_val"`
		NewFileValue schema.File `gorethink:"new_val"`
	}
	var fileDoc doc.File

	files, _ := r.Table("datafiles").Changes().Run(session)
	for files.Next(&change) {
		id := getFileID(change.OldFileValue, change.NewFileValue)
		indexer := search.NewSingleFileIndexer(client, session, id)
		indexer.Do("files", fileDoc)
	}
}

func getFileID(oldFile, newFile schema.File) string {
	if oldFile.ID != "" {
		return oldFile.ID
	}
	return newFile.ID
}

type sample struct {
	ID string `gorethink:"id"`
}

func sampleChangeIndexer(client *elastic.Client, session *r.Session) {
	var change struct {
		OldSampleValue sample `gorethink:"old_val"`
		NewSampleValue sample `gorethink:"new_val"`
	}
	var sampleDoc doc.Sample

	samples, _ := r.Table("samples").Changes().Run(session)
	for samples.Next(&change) {
		id := getSampleID(change.OldSampleValue, change.NewSampleValue)
		indexer := search.NewSingleSampleIndexer(client, session, id)
		indexer.Do("samples", sampleDoc)
	}
}

func getSampleID(oldSample, newSample sample) string {
	if oldSample.ID != "" {
		return oldSample.ID
	}
	return newSample.ID
}
