package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/filters"
	"gopkg.in/olivere/elastic.v2"
	"encoding/json"
)

type searchResource struct {
	log *app.Logger
}

type Query struct {
	QString string `json:"query_string"`
}

func newSearchResource() rest.Service {
	return &searchResource{
		log: app.NewLog("resource", "search"),
	}
}

func (r *searchResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/search").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)

	ws.Route(ws.GET("/project/{project}/files").
		Filter(filters.ProjectAccess).
		Filter(filters.SearchClient).
		To(rest.RouteHandler(r.searchProjectFiles)).
		Param(ws.PathParameter("project", "project id").DataType("string")).
		Doc("Searches files in project").
		Reads(Query{}).
		Writes([]schema.File{}))

	return ws
}

//func (r *searchResource) searchUserProjects(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
//	return nil, nil
//}

func (r *searchResource) searchProjectFiles(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var (
		query Query
		matches []schema.File
	)

	if err := request.ReadEntity(&query); err != nil {
		r.log.Debugf("Failed reading query: %s", err)
		return matches, err
	}

	project := request.Attribute("project").(schema.Project)
	client := request.Attribute("searchclient").(*elastic.Client)
	termQueryProj := elastic.NewTermQuery("project_id", project.ID)
	userQuery := elastic.NewQueryStringQuery(query.QString)
	boolQuery := elastic.NewBoolQuery()
	boolQuery = boolQuery.Must(termQueryProj, userQuery)
	results, err := client.Search().Index("mc").Type("files").Query(&boolQuery).Do()
	if err != nil {
		r.log.Infof("Query failed: %s", err)
		return matches, err
	}

	if results.Hits != nil {
		for _, hit := range results.Hits.Hits {
			var f schema.File
			if err := json.Unmarshal(*hit.Source, &f); err == nil {
				matches = append(matches, f)
			}
		}
	}

	return matches, nil
}
