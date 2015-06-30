package schema

import "time"

type Access struct {
	ID          string    `gorethink:"id,omitempty"`
	Dataset     string    `gorethink:"dataset"`
	Birthtime   time.Time `gorethink:"birthtime"`
	MTime       time.Time `gorethink:"mtime"`
	Permissions string    `gorethink:"permissions"`
	ProjectID   string    `gorethink:"project_id"`
	ProjectName string    `gorethink:"project_name"`
	Status      string    `gorethink:"status"`
	UserID      string    `gorethink:"user_id"`
}

func NewAccess(projectID, projectName, userID string) Access {
	now := time.Now()
	return Access{
		Birthtime:   now,
		MTime:       now,
		ProjectID:   projectID,
		ProjectName: projectName,
		UserID:      userID,
	}
}
