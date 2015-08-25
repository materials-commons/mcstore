package mcstore

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
	"gopkg.in/olivere/elastic.v2"
)

type idField struct {
	ID string `gorethink:"id"`
}

type changeItem struct {
	OldValue idField `gorethink:"old_val"`
	NewValue idField `gorethink:"new_val"`
}

func processChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change changeItem
	)

	processes, _ := r.Table("processes").Changes().Run(session)
	for processes.Next(&change) {
		id := getItemID(change.OldValue, change.NewValue)
		app.Log.Infof("Indexing process id: %s", id)
		indexProcesses(client, session, id)
	}
}

func fileChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change changeItem
	)

	files, _ := r.Table("datafiles").Changes().Run(session)
	for files.Next(&change) {
		id := getItemID(change.OldValue, change.NewValue)
		app.Log.Infof("Indexing file id: %s", id)
		indexFiles(client, session, id)
		indexSamplesUsingFile(client, session, id)
	}
}

type proj2fileChange struct {
	OldValue schema.Project2DataFile `gorethink:"old_val"`
	NewValue schema.Project2DataFile `gorethink:"new_val"`
}

func projectFileChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change proj2fileChange
	)

	files, _ := r.Table("project2datafile").Changes().Run(session)
	for files.Next(&change) {
		fileID := getFileID(change.OldValue, change.NewValue)
		app.Log.Infof("Indexing file added to project: %s", fileID)
		indexFiles(client, session, fileID)
	}
}

func getFileID(oldVal, newVal schema.Project2DataFile) string {
	if oldVal.ID != "" {
		return oldVal.DataFileID
	}
	return newVal.DataFileID
}

func sampleChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change changeItem
	)

	samples, _ := r.Table("samples").Changes().Run(session)
	for samples.Next(&change) {
		id := getItemID(change.OldValue, change.NewValue)
		app.Log.Infof("Indexing sample id: %s", id)
		indexSamples(client, session, id)
	}
}

type note2item struct {
	ItemType string `gorethink:"item_type"`
	ItemID   string `gorethink:"item_id"`
	NoteID   string `gorethink:"note_id"`
}

func noteChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change changeItem
		n2i    note2item
	)

	notes, _ := r.Table("notes").Changes().Run(session)
	for notes.Next(&change) {
		id := getItemID(change.OldValue, change.NewValue)
		rql := r.Table("note2item").GetAllByIndex("note_id", id)
		if err := model.GetRow(rql, session, &n2i); err != nil {
			app.Log.Errorf("noteChangeIndexer GetRow err: %s", err)
			continue
		}
		if n2i.ItemType == "datafile" {
			app.Log.Infof("Index datafile because of note: %s", n2i.ItemID)
			indexFiles(client, session, n2i.ItemID)
			indexSamplesUsingFile(client, session, n2i.ItemID)
		}
	}
}

type noteItemChange struct {
	OldValue note2item `gorethink:"old_val"`
	NewValue note2item `gorethink:"new_val"`
}

func noteItemChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change noteItemChange
	)

	items, _ := r.Table("note2item").Changes().Run(session)
	for items.Next(&change) {
		noteItem := getNoteItem(change)
		if noteItem.ItemType == "datafile" {
			app.Log.Infof("Index datafile because of note: %s", noteItem.ItemID)
			indexFiles(client, session, noteItem.ItemID)
			indexSamplesUsingFile(client, session, noteItem.ItemID)
		}
	}
}

func getNoteItem(change noteItemChange) note2item {
	if change.OldValue.ItemID != "" {
		return change.OldValue
	}
	return change.NewValue
}

type sample2datafile struct {
	SampleID   string `gorethink:"sample_id"`
	DataFileID string `gorethink:"datafile_id"`
}

func indexSamplesUsingFile(client *elastic.Client, session *r.Session, fileID string) {
	var (
		s2df schema.Sample2DataFile
	)

	if samples, err := r.Table("sample2datafile").GetAllByIndex("datafile_id", fileID).Run(session); err != nil {
		return
	} else {
		for samples.Next(&s2df) {
			indexSamples(client, session, s2df.SampleID)
		}
	}
}

type propertyset2property struct {
	ID            string `gorethink:"id"`
	PropertySetID string `gorethink:"property_set_id"`
}

type ps2pChange struct {
	OldValue propertyset2property `gorethink:"old_val"`
	NewValue propertyset2property `gorethink:"new_val"`
}

type sampleIDItem struct {
	SampleID string `gorethink:"sample_id"`
}

func propertysetChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change ps2pChange
		sample sampleIDItem
	)

	psets, _ := r.Table("propertyset2property").Changes().Run(session)
	for psets.Next(&change) {
		psetID := getPropertySetID(change.OldValue, change.NewValue)
		rql := r.Table("sample2propertyset").GetAllByIndex("property_set_id", psetID)
		if err := model.GetRow(rql, session, &sample); err != nil {
			app.Log.Errorf("propertysetChangeIndexer GetRow err: %s", err)
			continue
		}
		app.Log.Infof("Index sample because of property change: %s", sample.SampleID)
		indexSamples(client, session, sample.SampleID)
	}
}

func getPropertySetID(oldPS, newPS propertyset2property) string {
	if oldPS.ID != "" {
		return oldPS.PropertySetID
	}
	return newPS.PropertySetID
}

type s2dfChange struct {
	OldValue sampleIDItem `gorethink:"old_val"`
	NewValue sampleIDItem `gorethink:"new_val"`
}

func sampleDatafileChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change s2dfChange
	)

	sampleFiles, _ := r.Table("sample2datafile").Changes().Run(session)
	for sampleFiles.Next(&change) {
		id := getSampleID(change.OldValue, change.NewValue)
		app.Log.Infof("Index sample because of file change: %s", id)
		indexSamples(client, session, id)
	}
}

func getSampleID(oldItem, newItem sampleIDItem) string {
	if oldItem.SampleID != "" {
		return oldItem.SampleID
	}
	return newItem.SampleID
}

type tag2Item struct {
	ItemType string `gorethink:"item_type"`
	ItemID   string `gorethink:"item_id"`
}

type tag2ItemChange struct {
	OldValue tag2Item `gorethink:"old_val"`
	NewValue tag2Item `gorethink:"new_val"`
}

func tagChangeIndexer(client *elastic.Client, session *r.Session) {
	var (
		change tag2ItemChange
	)

	tagUpdates, _ := r.Table("tag2item").Changes().Run(session)
	for tagUpdates.Next(&change) {
		itemID, itemType := getTagItemIDAndType(change.OldValue, change.NewValue)
		if itemType == "datafile" {
			app.Log.Infof("Index file because of tag change: %s", itemID)
			indexFiles(client, session, itemID)
			indexSamplesUsingFile(client, session, itemID)
		}
	}
}

func getTagItemIDAndType(oldTagItem, newTagItem tag2Item) (itemID, itemType string) {
	if oldTagItem.ItemID != "" {
		return oldTagItem.ItemID, oldTagItem.ItemType
	}
	return newTagItem.ItemID, newTagItem.ItemType
}

func getItemID(oldItem, newItem idField) string {
	if oldItem.ID != "" {
		return oldItem.ID
	}
	return newItem.ID
}

func indexFiles(client *elastic.Client, session *r.Session, fileIDs ...interface{}) {
	var fileDoc doc.File
	indexer := search.NewMultiFileIndexer(client, session, fileIDs...)
	indexer.Do("files", fileDoc)
}

func indexSamples(client *elastic.Client, session *r.Session, sampleIDs ...interface{}) {
	var sampleDoc doc.Sample
	indexer := search.NewMultiSampleIndexer(client, session, sampleIDs...)
	indexer.Do("samples", sampleDoc)
}

func indexProcesses(client *elastic.Client, session *r.Session, processIDs ...interface{}) {
	var processDoc doc.Process
	indexer := search.NewMultiProcessIndexer(client, session, processIDs...)
	indexer.Do("processes", processDoc)
}
