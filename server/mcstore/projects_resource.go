package mcstore

import (
	"archive/zip"
	"net/http"
	"os"
	"path/filepath"

	"io/ioutil"

	rethinkdb "github.com/dancannon/gorethink"
	"github.com/emicklei/go-restful"
	"github.com/hashicorp/go-uuid"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/materials-commons/mcstore/pkg/ws/rest"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
)

// An projectsResource holds the state and services needed for the
// projects REST resource.
type projectsResource struct {
	log *app.Logger
}

// newProjectsResource creates a new projects resource.
func newProjectsResource() *projectsResource {
	return &projectsResource{
		log: app.NewLog("resource", "projects"),
	}
}

// WebService creates an instance of the projects web service.
func (r *projectsResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/project2").Produces(restful.MIME_JSON).Consumes(restful.MIME_JSON)

	ws.Route(ws.POST("").To(rest.RouteHandler(r.createProject)).
		Doc("Creates a new project for user. If project exists it returns the existing project.").
		Reads(mcstoreapi.CreateProjectRequest{}).
		Writes(mcstoreapi.CreateProjectResponse{}))

	ws.Route(ws.POST("directory").To(rest.RouteHandler(r.getDirectory)).
		Doc("Gets or creates a directory by its directory path").
		Reads(mcstoreapi.GetDirectoryRequest{}).
		Writes(mcstoreapi.GetDirectoryResponse{}))

	ws.Route(ws.POST("archive").To(rest.RouteHandler(r.createDownloadZipFile)).
		Doc("Creates a zipfile archive of request files ids"))

	ws.Route(ws.GET("/download/archive/{archive}").To(rest.RouteHandler1(r.downloadArchiveZipFile)).
		Doc("Download a created archive"))

	return ws
}

// createProject services the create project request. It will ensure that the user
// doesn't have a project matching the given project name.
func (r *projectsResource) createProject(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*rethinkdb.Session)
	var req mcstoreapi.CreateProjectRequest
	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("createProject ReadEntity failed: %s", err)
		return nil, err
	}

	projectService := newProjectService(session)
	proj, existing, err := projectService.createProject(req.Name, user.ID, req.MustNotExist)
	switch {
	case err != nil:
		return nil, err
	default:
		resp := &mcstoreapi.CreateProjectResponse{
			ProjectID: proj.ID,
			Existing:  existing,
		}
		return resp, nil
	}
}

// getDirectory services request to get a directory for a project. It accepts directories
// by their path relative to the project. The getDirectory service will create a directory
// that doesn't exist.
func (r *projectsResource) getDirectory(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	session := request.Attribute("session").(*rethinkdb.Session)
	var req mcstoreapi.GetDirectoryRequest
	if err := request.ReadEntity(&req); err != nil {
		app.Log.Debugf("getDirectory ReadEntity failed: %s", err)
		return nil, err
	}

	dirService := newDirService(session)
	dir, err := dirService.createDir(req.ProjectID, req.Path)
	switch {
	case err != nil:
		app.Log.Debugf("dirService.createDir failed for project/dir %s/%s: %s", req.ProjectID, req.Path, err)
		return nil, err
	default:
		resp := &mcstoreapi.GetDirectoryResponse{
			DirectoryID: dir.ID,
			Path:        req.Path,
		}
		return resp, nil
	}
}

func (r *projectsResource) createDownloadZipFile(request *restful.Request, response *restful.Response, user schema.User) (interface{}, error) {
	var (
		zipResponse struct {
			ArchiveID string `json:"archive_id"`
		}

		zipRequest struct {
			FileIDs []string `json:"file_ids"`
		}
	)

	if err := request.ReadEntity(&zipRequest); err != nil {
		app.Log.Debugf("createDownloadZipFile ReadEntity failed: %s", err)
		return nil, err
	}

	session := request.Attribute("session").(*rethinkdb.Session)

	u, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	zipfile, err := os.Create(filepath.Join("/tmp", u+".zip"))
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	rfiles := dai.NewRFiles(session)
	rprojects := dai.NewRProjects(session)
	rusers := dai.NewRUsers(session)
	access := domain.NewAccess(rprojects, rfiles, rusers)

	for _, fileID := range zipRequest.FileIDs {
		if f, err := access.GetFile(user.APIKey, fileID); err == nil {
			fpath := app.MCDir.FilePath(fileID)
			if contents, err := ioutil.ReadFile(fpath); err == nil {
				if f, err := archive.Create(f.Name); err == nil {
					f.Write(contents)
				}
			}
		}
	}

	zipResponse.ArchiveID = u

	return &zipResponse, nil
}

func (r *projectsResource) downloadArchiveZipFile(request *restful.Request, response *restful.Response, user schema.User) error {
	archiveZipPath := filepath.Join("/tmp", request.PathParameter("archive"))
	http.ServeFile(response.ResponseWriter, request.Request, archiveZipPath)
	defer os.Remove(archiveZipPath)
	return nil
}
