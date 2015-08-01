package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
	"gopkg.in/olivere/elastic.v2"
)

// NewServicesContainer creates a new restful.Container made up of all
// the rest resources handled by the server.
func NewServicesContainer(sc db.SessionCreater) *restful.Container {
	container := restful.NewContainer()

	databaseSessionFilter := &databaseSessionFilter{
		session: sc.RSession,
	}
	container.Filter(databaseSessionFilter.Filter)

	apikeyFilter := newAPIKeyFilter(apiKeyCache)
	container.Filter(apikeyFilter.Filter)

	// launch routine to track changes to users and
	// update the keycache appropriately.
	go updateKeyCacheOnChange(sc.RSessionMust(), apiKeyCache)

	// launch routines to monitor for database changes
	launchSearchIndexChangeMonitors(sc)

	uploadResource := newUploadResource()
	container.Add(uploadResource.WebService())

	projectsResource := newProjectsResource()
	container.Add(projectsResource.WebService())

	searchResource := newSearchResource()
	container.Add(searchResource.WebService())

	return container
}

func launchSearchIndexChangeMonitors(sc db.SessionCreater) {
	esclient := esClientMust()
	session := sc.RSessionMust()

	go processChangeIndexer(esclient, session)
	go fileChangeIndexer(esclient, session)
	go sampleChangeIndexer(esclient, session)
	go noteChangeIndexer(esclient, session)
	go propertysetChangeIndexer(esclient, session)
	go sampleDatafileChangeIndexer(esclient, session)
}

func esClientMust() *elastic.Client {
	url := esURL()
	app.Log.Infof("Connecting to search url: %s", url)
	c, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		app.Log.Panicf("Couldn't connect to ElasticSearch")
	}
	return c
}

func esURL() string {
	if esURL := config.GetString("MC_ES_URL"); esURL != "" {
		return esURL
	}
	return "http://localhost:9200"
}
