package main

import (
	"fmt"

	"os"

	"github.com/codegangsta/cli"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search"
	"github.com/materials-commons/mcstore/server/mcstore/pkg/search/doc"
	"gopkg.in/olivere/elastic.v2"
)

var mappings string = `
{
	"mappings": {
	     "files": {
	         "properties": {
	              "project_id": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              },
	              "project": {
	              		"type": "string",
	              		"index": "not_analyzed"
	              },
	              "datadir_id": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              },
	              "id": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              },
	              "usesid": {
	                  "type": "string",
	                  "index": "not_analyzed"
	              }
	         }
	     },
	     "projects": {
	     	"properties": {
	     		"id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	             },
	             "datadir": {
	             	"type": "string",
	             	"index": "not_analyzed"
	             }
	     	}
	     },
	     "samples": {
	     	"properties":{
	     	    "id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	            },
	        	"project_id": {
	                "type": "string",
	                "index": "not_analyzed"
	        	},
	        	"sample_id": {
	        		"type": "string",
	        		"index": "not_analyzed"
	        	}
	        }
	     },
	     "processes": {
	     	"properties":{
	     		"process_id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	             },
	             "project_id": {
	             	"type": "string",
	             	"index": "not_analyzed"
	             }
	        }
	     },
		"users": {
			"properties":{
	     		"id": {
	            	"type": "string",
	             	"index": "not_analyzed"
	             }
	        }
	    }
	}
}
`

func main() {
	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Authors = []cli.Author{
		{
			Name:  "V. Glenn Tarcea",
			Email: "gtarcea@umich.edu",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "es-url",
			Value:  "http://localhost:9500",
			Usage:  "Elasticsearch server URL",
			EnvVar: "MC_ES_URL",
		},

		cli.StringFlag{
			Name:   "db-connection",
			Value:  "localhost:30815",
			Usage:  "RethinkDB connection string",
			EnvVar: "MCDB_CONNECTION",
		},

		cli.StringFlag{
			Name:   "db-name",
			Value:  "materialscommons",
			Usage:  "Database to index",
			EnvVar: "MCDB_NAME",
		},

		cli.StringFlag{
			Name:   "mc-dir",
			Value:  "/mcfs/data/test",
			Usage:  "Path to data directory",
			EnvVar: "MCDIR",
		},

		cli.BoolFlag{
			Name:  "create-index",
			Usage: "Whether the index should be recreated",
		},

		cli.BoolTFlag{
			Name:  "processes",
			Usage: "Index processes",
		},

		cli.BoolTFlag{
			Name:  "files",
			Usage: "Index files",
		},

		cli.BoolTFlag{
			Name:  "samples",
			Usage: "Index samples",
		},

		cli.BoolTFlag{
			Name:  "projects",
			Usage: "Index projects",
		},

		cli.BoolTFlag{
			Name:  "users",
			Usage: "Index users",
		},
	}

	app.Action = mcbulkCLI
	app.Run(os.Args)
}

func mcbulkCLI(c *cli.Context) {
	setupConfig(c)
	runCommands(c)
}

func setupConfig(c *cli.Context) {
	esurl := c.String("es-url")
	config.Set("MC_ES_URL", esurl)

	dbname := c.String("db-name")
	config.Set("MCDB_NAME", dbname)

	dbcon := c.String("db-connection")
	config.Set("MCDB_CONNECTION", dbcon)
}

func runCommands(c *cli.Context) {
	esurl := esURL()
	fmt.Println("Elasticsearch URL:", esurl)
	client, err := elastic.NewClient(elastic.SetURL(esurl))
	if err != nil {
		panic("Unable to connect to elasticsearch")
	}

	session := db.RSessionMust()

	if c.Bool("create-index") {
		createIndex(client)
	}

	if c.BoolT("files") {
		loadFiles(client, session)
	}

	if c.BoolT("users") {
		loadUsers(client, session)
	}

	if c.BoolT("projects") {
		loadProjects(client, session)
	}

	if c.BoolT("samples") {
		loadSamples(client, session)
	}

	if c.BoolT("processes") {
		loadProcesses(client, session)
	}
}

func esURL() string {
	if esURL := config.GetString("MC_ES_URL"); esURL != "" {
		return esURL
	}
	return "http://localhost:9200"
}

func createIndex(client *elastic.Client) {
	fmt.Println("Creating index mc...")

	exists, err := client.IndexExists("mc").Do()
	if err != nil {
		panic("  Failed checking index existence")
	}

	if exists {
		fmt.Println("  Index exists deleting old one")
		client.DeleteIndex("mc").Do()
	}

	createStatus, err := client.CreateIndex("mc").Body(mappings).Do()
	if err != nil {
		fmt.Println("  Failed creating index: ", err)
		os.Exit(1)
	}

	if !createStatus.Acknowledged {
		fmt.Println("  Index create not acknowledged")
	}

	fmt.Println("Done.")
}

func loadFiles(client *elastic.Client, session *r.Session) {
	var df doc.File
	filesIndexer := search.NewFilesIndexer(client, session)

	fmt.Println("Indexing files...")
	if err := filesIndexer.Do("files", df); err != nil {
		fmt.Println("  Indexing files failed:", err)
		fmt.Println("  Some files may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadUsers(client *elastic.Client, session *r.Session) {
	var u schema.User
	usersIndexer := search.NewUsersIndexer(client, session)

	fmt.Println("Indexing users...")
	if err := usersIndexer.Do("users", u); err != nil {
		fmt.Println("  Indexing users failed:", err)
		fmt.Println("  Some users may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadProjects(client *elastic.Client, session *r.Session) {
	var p schema.Project
	projectsIndexer := search.NewProjectsIndexer(client, session)

	fmt.Println("Indexing projects...")
	if err := projectsIndexer.Do("projects", p); err != nil {
		fmt.Println("  Indexing projects failed:", err)
		fmt.Println("  Some projects may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadSamples(client *elastic.Client, session *r.Session) {
	var sample doc.Sample
	samplesIndexer := search.NewSamplesIndexer(client, session)

	fmt.Println("Indexing samples...")
	if err := samplesIndexer.Do("samples", sample); err != nil {
		fmt.Println("  Indexing samples failed:", err)
		fmt.Println("  Some samples may not have been indexed.")
	}
	fmt.Println("Done.")
}

func loadProcesses(client *elastic.Client, session *r.Session) {
	var process doc.Process
	processesIndexer := search.NewProcessesIndexer(client, session)

	fmt.Println("Indexing processes...")
	if err := processesIndexer.Do("processes", process); err != nil {
		fmt.Println("  Indexing processes failed:", err)
		fmt.Println("  Some processes may not have been indexed.")
	}
	fmt.Println("Done.")
}
