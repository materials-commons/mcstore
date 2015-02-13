package project

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type dirTree struct {
	allDirs      map[string]*Node
	topLevelDirs []Node
}

func newDirTree() *dirTree {
	return &dirTree{
		allDirs: make(map[string]*Node),
	}
}

func (t *dirTree) build(dirs []schema.Directory) []Node {
	for _, dir := range dirs {
		dirNode := t.createDirNode(dir)
		t.addFiles(dirNode, &dir)
		t.addToParent(dirNode)
	}
	return t.topLevelDirs
}

func (t *dirTree) createDirNode(dir schema.Directory) *Node {
	dirNode := &Node{
		ID:        dir.ID,
		Name:      dir.Name,
		Type:      "datadir",
		Owner:     dir.Owner,
		Birthtime: dir.Birthtime,
		Size:      0,
		Level:     strings.Count(dir.Name, "/"),
	}

	// The item may have been added as a parent
	// before it was actually seen. We check for
	// this case and grab the children to add to
	// us now that we have the details for the ditem.
	if existingDir, found := t.allDirs[dir.Name]; found {
		dirNode.Children = existingDir.Children
	}
	t.allDirs[dir.Name] = dirNode

	// Is this a root level directory?
	if dirNode.Level == 0 {
		t.topLevelDirs = append(t.topLevelDirs, *dirNode)
	}
	return dirNode
}

func (t *dirTree) addFiles(dirNode *Node, dir *schema.Directory) {
	for _, file := range dir.Files {
		if skipFile(&file) {
			continue
		}
		fileNode := createFileNode(dir.Name, &file)
		dirNode.Children = append(dirNode.Children, fileNode)
	}
}

func skipFile(file *schema.File) bool {
	switch {
	case ignoredFile(file):
		return true
	case !file.Current:
		return true
	default:
		return false
	}
}

func ignoredFile(file *schema.File) bool {
	// Only ignore dot files for now.
	return string(file.Name[0]) == "."
}

func createFileNode(dirName string, file *schema.File) Node {
	fileNode := Node{
		ID:        file.ID,
		Name:      file.Name,
		Type:      "datafile",
		Owner:     file.Owner,
		Birthtime: file.Birthtime,
		Size:      file.Size,
	}
	fileNode.Fullname = dirName + "/" + file.Name
	if file.MediaType.Mime == "" {
		fileNode.MediaType = "unknown"
	} else {
		fileNode.MediaType = file.MediaType.Mime
	}
	return fileNode
}

func (t *dirTree) addToParent(dirNode *Node) {
	parentName := filepath.Dir(dirNode.Name)
	if parentNode, found := t.allDirs[parentName]; found {
		parentNode.Children = append(parentNode.Children, *dirNode)
	} else {
		// We haven't seen the parent yet, but we need
		// to add the children. So, create a parent with
		// name and add children. When we finally see it
		// we will grab the children and add them to the
		// real object.
		parent := &Node{
			Name: parentName,
			Type: "datadir",
		}
		parent.Children = append(parent.Children, *dirNode)
		t.allDirs[parentName] = parent
	}
}
