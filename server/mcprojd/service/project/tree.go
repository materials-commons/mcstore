package project

import (
	"path/filepath"
	"strings"
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
	Children    []*Node   `json:"children"`
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

func (s *treeService) Tree(req *TreeRequest) ([]*Node, error) {
	files, err := s.projects.Files(req.ProjectID)
	if err != nil {
		return nil, nil
	}
	allDirs := make(map[string]*Node)
	var topLevelDirs []*Node
	var existingDir *Node
	for _, dir := range files {
		node := &Node{
			ID:        dir.ID,
			Name:      dir.Name,
			Type:      "datadir",
			Owner:     dir.Owner,
			Birthtime: dir.Birthtime,
			Size:      0,
			Level:     strings.Count(dir.Name, "/"),
		}
		if n, found := allDirs[dir.Name]; found {
			existingDir = n
			node.Children = existingDir.Children
		}
		if node.Level == 0 {
			topLevelDirs = append(topLevelDirs, node)
		}
		for _, file := range dir.Files {
			if string(file.Name[0]) == "." {
				continue
			}
			if !file.Current {
				continue
			}
			fileNode := &Node{
				ID:        file.ID,
				Name:      file.Name,
				Type:      "datafile",
				Owner:     file.Owner,
				Birthtime: file.Birthtime,
				Size:      file.Size,
			}
			fileNode.Fullname = dir.Name + "/" + file.Name
			if file.MediaType.Mime == "" {
				fileNode.MediaType = "unknown"
			} else {
				fileNode.MediaType = file.MediaType.Mime
			}
			node.Children = append(node.Children, fileNode)
		}
		parentName := filepath.Dir(node.Name)
		if n, found := allDirs[parentName]; found {
			n.Children = append(n.Children, node)
		} else {
			parent := &Node{
				Name: parentName,
				Type: "datadir",
			}
			parent.Children = append(parent.Children, node)
			allDirs[parentName] = parent
		}
	}
	return topLevelDirs, nil
}
