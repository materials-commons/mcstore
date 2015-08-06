package mcstore

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/filters"
	"gopkg.in/olivere/elastic.v2"
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

	ws.Route(ws.POST("/project/{project}").
		Filter(filters.ProjectAccess).
		Filter(filters.SearchClient).
		To(rest.RouteHandler(r.searchProject)).
		Param(ws.PathParameter("project", "project id").DataType("string")).
		Doc("Searches all project items - files, processes, samples, notes, users").
		Reads(Query{}).
		Writes(elastic.SearchHits{}))

	ws.Route(ws.POST("/project/{project}/files").
		Filter(filters.ProjectAccess).
		Filter(filters.SearchClient).
		To(rest.RouteHandler(r.searchProjectFiles)).
		Param(ws.PathParameter("project", "project id").DataType("string")).
		Doc("Searches files in project").
		Reads(Query{}).
		Writes(elastic.SearchHits{}))

	return ws
}

func (r *searchResource) searchProject(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var query Query
	if err := request.ReadEntity(&query); err != nil {
		r.log.Debugf("Failed reading query: %s", err)
		return nil, err
	}

	project := request.Attribute("project").(schema.Project)
	client := request.Attribute("searchclient").(*elastic.Client)
	q := createQuery(query, project.ID)
	results, err := client.Search().Index("mc").Query(q).Size(100).Do()
	if err != nil {
		r.log.Infof("Query failed: %s", err)
		return nil, err
	}
	return results.Hits, nil
}

//func (r *searchResource) searchUserProjects(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
//	return nil, nil
//}

func (r *searchResource) searchProjectFiles(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var query Query

	if err := request.ReadEntity(&query); err != nil {
		r.log.Debugf("Failed reading query: %s", err)
		return nil, err
	}

	project := request.Attribute("project").(schema.Project)
	client := request.Attribute("searchclient").(*elastic.Client)
	q := createQuery(query, project.ID)
	results, err := client.Search().Index("mc").Type("files,processes,samples").Query(q).Size(100).Do()
	if err != nil {
		r.log.Infof("Query failed: %s", err)
		return nil, err
	}

	return results.Hits, nil
}

func createQuery(query Query, projectID string) elastic.Query {
	termFilterProj := elastic.NewTermFilter("project_id", projectID)
	userQuery := elastic.NewQueryStringQuery(query.QString)
	boolTerm := elastic.NewBoolFilter()
	boolTerm = boolTerm.Must(termFilterProj, userQuery)
	return &boolTerm
}
