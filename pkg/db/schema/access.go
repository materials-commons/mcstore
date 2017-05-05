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

type AccessPermission string

const (
	PermReadWrite       AccessPermission = "RW"
	PermRead            AccessPermission = "R"
	PermReadWriteDelete AccessPermission = "RWD"
)

type Access2 struct {
	ID          string           `gorethink:"id,omitempty"`
	Owner       string           `gorethink:"owner"`
	GroupName   string           `gorethink:"group_name"`
	Permissions AccessPermission `gorethink:"permissions"`
	Users       []string         `gorethink:"users"`
	Items       []Access2Item    `gorethink:"items"`
}

type Access2Item struct {
	ID       string `gorethink:"id,omitempty"`
	ItemType string `gorethink:"item_type"`
	ItemID   string `gorethink:"item_id"`
}

func Owner(owner string) func(a *Access2) {
	return func(a *Access2) {
		a.Owner = owner
	}
}

func Users(users ...string) func(a *Access2) {
	return func(a *Access2) {
		a.Users = append(a.Users, users...)
	}
}

func Permission(perm AccessPermission) func(a *Access2) {
	return func(a *Access2) {
		a.Permissions = perm
	}
}

func PermittedItem(itemID, itemType string) func(a *Access2) {
	return func(a *Access2) {
		a.Items = append(a.Items, Access2Item{ItemID: itemID, ItemType: itemType})
	}
}

func NewAccess2(owner string, name string) {

}
