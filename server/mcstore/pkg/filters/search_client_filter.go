package filters

import (
	"net/http"

	"sync"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"gopkg.in/olivere/elastic.v2"
)

var (
	clientInit sync.Once
	client     *elastic.Client
)

func SearchClient(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if client := getSearchClient(); client == nil {
		response.WriteErrorString(http.StatusInternalServerError, "Unable to connect to search service")
	} else {
		request.SetAttribute("searchclient", client)
		chain.ProcessFilter(request, response)
	}
}

func getSearchClient() *elastic.Client {
	clientInit.Do(func() {
		url := esURL()
		c, err := elastic.NewClient(elastic.SetURL(url))
		if err != nil {
			app.Log.Errorf("Couldn't connect to ElasticSearch")
		}
		client = c
	})
	return client
}

func esURL() string {
	if esURL := config.GetString("MC_ES_URL"); esURL != "" {
		return esURL
	}
	return "http://localhost:9200"
}
