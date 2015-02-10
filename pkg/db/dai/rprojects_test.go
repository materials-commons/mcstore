package dai

import (
	"testing"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/db"
)

func TestRProjectsFiles(t *testing.T) {
	config.Set("MCDB_CONNECTION", "localhost:30815")
	config.Set("MCDB_NAME", "materialscommons")
	var rprojects = NewRProjects(db.RSessionMust())
	rprojects.Files("e95944cb-2bfc-4d56-8be6-72029fd4d1ad")
	//fmt.Println(err)
	//fmt.Printf("%#v\n", dirs)
}
