package uploads

import "time"

type CreateRequest struct {
	User        string
	DirectoryID string
	ProjectID   string
	Host        string
	Birthtime   time.Time
}
