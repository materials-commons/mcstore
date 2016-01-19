package mc

import (
	"os"
	"path/filepath"
	"strings"

	"time"

	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type downloader struct {
	projectDB  ProjectDB
	dir        *Directory
	file       *File
	serverFile *schema.File
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
		return ErrFileNotUploaded
	case finfo.ModTime().Unix() > d.file.MTime.Unix():
		// Existing file with updates that haven't been uploaded. Don't overwrite.
		return ErrFileVersionNotUploaded
	default:
		return d.downloadNewFile(path)
	}
}

func pathFromProject(path, projectName string) string {
	index := strings.Index(path, projectName)
	return path[index:len(path)]
}
