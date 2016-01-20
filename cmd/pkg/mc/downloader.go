package mc

import (
	"os"
	"path/filepath"
	"strings"

	"time"

	"fmt"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
)

type downloader struct {
	projectDB  ProjectDB
	dir        *Directory
	file       *File
	serverFile *mcstoreapi.ServerFile
	c          *ClientAPI
}

func newDownloader(projectDB ProjectDB, clientAPI *ClientAPI) *downloader {
	return &downloader{
		projectDB: projectDB,
		c:         clientAPI,
	}
}

func (d *downloader) downloadFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != err {
		return err
	}
	project := d.projectDB.Project()
	var err error
	if d.serverFile, err = d.c.serverAPI.GetFileForPath(project.ProjectID, pathFromProject(path, project.Name)); err != nil {
		fmt.Println("getFileForPath returned error", err)
		return err
	}
	if d.dir, err = d.projectDB.FindDirectory(filepath.Dir(path)); err != nil {
		d.c.createDirectory(d.projectDB, filepath.Dir(path))
		d.dir, _ = d.projectDB.FindDirectory(filepath.Dir(path))
	}

	d.file, _ = d.projectDB.FindFile(filepath.Base(path), d.dir.ID)

	if finfo, err := os.Stat(path); os.IsNotExist(err) {
		return d.downloadNewFile(path)
	} else {
		return d.downloadExistingFile(finfo, path)
	}
}

func (d *downloader) downloadNewFile(path string) error {
	project := d.projectDB.Project()
	if err := d.c.serverAPI.DownloadFile(project.ProjectID, d.serverFile.ID, path); err != nil {
		fmt.Println("serverAPI.DownloadFile error", err)
		return err
	}
	finfo, _ := os.Stat(path)
	if d.file == nil {
		// File not in database

		newFile := File{
			FileID:     d.serverFile.ID,
			Name:       filepath.Base(path),
			Checksum:   d.serverFile.Checksum,
			Size:       finfo.Size(),
			MTime:      finfo.ModTime(),
			LastUpload: time.Now(),
			Directory:  d.dir.ID,
		}
		d.projectDB.InsertFile(&newFile)
	} else {
		// Existing file but not on users system
		d.file.MTime = finfo.ModTime()
		d.file.LastUpload = time.Now()
		d.file.FileID = d.serverFile.ID
		d.file.Size = finfo.Size()
		d.file.Checksum = d.serverFile.Checksum
		d.projectDB.UpdateFile(d.file)
	}
	return nil
}

func (d *downloader) downloadExistingFile(finfo os.FileInfo, path string) error {
	switch {
	case d.file == nil:
		// There is an existing file that isn't in database. Don't overwrite.
		fmt.Println("downloadExistingFile ErrFileNotUploaded")
		return ErrFileNotUploaded
	case finfo.ModTime().Unix() > d.file.MTime.Unix():
		// Existing file with updates that haven't been uploaded. Don't overwrite.
		fmt.Println("downloadExistingFile ErrFileVersionNotUploaded")
		return ErrFileVersionNotUploaded
	case d.file.Checksum == d.serverFile.Checksum:
		// Latest file already downloaded
		fmt.Println("downloadExistingFile Checksums are equal")
		return nil
	default:
		return d.downloadNewFile(path)
	}
}

func pathFromProject(path, projectName string) string {
	index := strings.Index(path, projectName)
	return path[index:len(path)]
}

type fentry struct {
	Type     string
	ID       string
	Path     string
	Size     int64
	Checksum string
}

type projectDownloader struct {
	downloader *downloader
	files      []fentry
}

func newProjectDownloader(projectDB ProjectDB, clientAPI *ClientAPI) *projectDownloader {
	return &projectDownloader{
		downloader: newDownloader(projectDB, clientAPI),
		files:      []fentry{},
	}
}

func (d *projectDownloader) downloadProject() error {
	project := d.downloader.projectDB.Project()

	if dir, err := d.downloader.c.getProjectDirList(project.ProjectID, ""); err == nil {
		d.files = append(d.files, toFentry(dir))
		d.loadDirRecurse(project.ProjectID, dir)
	}

	// Project Path contains the name of the project. The path for each entry
	// start with the project name. So we remove the project name from the
	// project path since the entry path will contain it.
	// eg, project path: /home/me/projects/PROJECT_NAME
	// individual entry paths: PROJECT_NAME/myfile.txt
	// so projectDir removes PROJECT_NAME, since joining with entry.Path will
	// put the PROJECT_NAME back in to the path.
	projectDir := filepath.Dir(project.Path)
	for _, e := range d.files {
		if e.Type == "file" {
			fmt.Println("Downloading file to:", filepath.Join(projectDir, e.Path))
			d.downloader.downloadFile(filepath.Join(projectDir, e.Path))
		} else if e.Type == "directory" {
			fmt.Println("Creating directory:", filepath.Join(projectDir, e.Path))
			d.createDir(filepath.Join(projectDir, e.Path), e.ID)
		}
	}
	return nil
}

func (d *projectDownloader) loadDirRecurse(projectID string, dentry *mcstoreapi.ServerDir) {
	if dentry.Type == "directory" {
		if dir, err := d.downloader.c.getProjectDirList(projectID, dentry.ID); err == nil {
			for _, entry := range dir.Children {
				d.files = append(d.files, toFentry(&entry))
				d.loadDirRecurse(projectID, &entry)
			}
		}
	}
}

func (d *projectDownloader) createDir(path, dirID string) {
	os.MkdirAll(path, 0770)
	if _, err := d.downloader.projectDB.FindDirectory(path); err == app.ErrNotFound {
		dir := &Directory{
			DirectoryID: dirID,
			Path:        path,
		}
		d.downloader.projectDB.InsertDirectory(dir)
	}
}

func toFentry(dentry *mcstoreapi.ServerDir) fentry {
	entry := fentry{
		Type:     dentry.Type,
		ID:       dentry.ID,
		Path:     dentry.Path,
		Size:     dentry.Size,
		Checksum: dentry.Checksum,
	}
	return entry
}
