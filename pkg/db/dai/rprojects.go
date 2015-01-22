package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type rProjects struct {
	session *r.Session
}

func NewRProjects(session *r.Session) rProjects {
	return rProjects{
		session: session,
	}
}

func (p rProjects) ByID(id string) (*schema.Project, error) {
	var project schema.Project
	if err := model.Projects.Qs(p.session).ByID(id, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

func (p rProjects) HasDirectory(projectID, dirID string) bool {
	rql := model.ProjectDirs.T().GetAllByIndex("datadir_id", dirID)
	var proj2dir []schema.Project2DataDir
	if err := model.ProjectDirs.Qs(p.session).Rows(rql, &proj2dir); err != nil {
		return false
	}

	// Look for matching projectID
	for _, entry := range proj2dir {
		if entry.ProjectID == projectID {
			return true
		}
	}

	return false
}
