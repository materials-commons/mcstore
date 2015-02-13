package project

import (
	"time"

	"github.com/materials-commons/mcstore/pkg/db"
	"github.com/materials-commons/mcstore/pkg/db/dai"
)

type TreeRequest struct {
	ProjectID string
}

type Node struct {
	ID          string    `json:"id"`
	Selected    bool      `json:"selected"`
	ParentID    string    `json:"parent_id"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	MediaType   string    `json:"mediatype"`
	Birthtime   time.Time `json:"birthtime"`
	Size        int64     `json:"size"`
	DisplayName string    `json:"displayname"`
	Fullname    string    `json:"fullname"`
	Type        string    `json:"type"`
	Children    []Node    `json:"children"`
	Level       int       `json:"level"`
	Tags        []string  `json:"tags"`
}

type TreeService interface {
	Tree(req *TreeRequest) ([]*Node, error)
}

type treeService struct {
	files    dai.Files
	projects dai.Projects
}

func NewTreeService() *treeService {
	session := db.RSessionMust()
	return &treeService{
		files:    dai.NewRFiles(session),
		projects: dai.NewRProjects(session),
	}
}

func (s *treeService) Tree(req *TreeRequest) ([]Node, error) {
	files, err := s.projects.Files(req.ProjectID)
	if err != nil {
		return nil, nil
	}
	dirTree := newDirTree()
	return dirTree.build(files), nil
}
