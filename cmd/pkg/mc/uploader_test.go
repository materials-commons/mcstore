package mc

import (
	"net/http/httptest"
	"time"

	"os"
	"path/filepath"

	"fmt"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/files"
	"github.com/materials-commons/mcstore/pkg/testdb"
	"github.com/materials-commons/mcstore/server/mcstore"
	"github.com/materials-commons/mcstore/server/mcstore/mcstoreapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = fmt.Println

var _ = Describe("ProjectUploader", func() {
	var (
		api           *mcstoreapi.ServerAPI
		server        *httptest.Server
		container     *restful.Container
		rr            *httptest.ResponseRecorder
		uploads       dai.Uploads
		uploadRequest mcstoreapi.CreateUploadRequest
		u             *uploader
	)

	const mcdirPath = "/tmp/mcdir"
	const projPath = "/tmp/test"

	BeforeEach(func() {
		config.Set("MCDIR", mcdirPath)
		container = mcstore.NewServicesContainer(testdb.Sessions)
		server = httptest.NewServer(container)
		rr = httptest.NewRecorder()
		config.Set("mcurl", server.URL)
		config.Set("apikey", "test")
		uploads = dai.NewRUploads(testdb.RSessionMust())
		api = mcstoreapi.NewServerAPI()
		uploadRequest = mcstoreapi.CreateUploadRequest{
			ProjectID:   "test",
			DirectoryID: "test",
			FileName:    "testreq.txt",
			FileSize:    4,
			ChunkSize:   2,
			FileMTime:   time.Now().Format(time.RFC1123),
			Checksum:    "abc123",
		}
		projectOpener := sqlProjectDBOpener{
			configer: configConfiger{},
		}
		config.Set("mcconfigdir", ".materialscommons")
		os.Mkdir(".materialscommons", 0777)
		projectDBSpec := ProjectDBSpec{
			Path:      projPath,
			Name:      "test",
			ProjectID: "test",
		}
		pdb, err := projectOpener.CreateProjectDB(projectDBSpec)
		if err == app.ErrExists {
			pdb, err = projectOpener.OpenProjectDB("test")
		}
		Expect(err).To(BeNil())
		u = newUploader(pdb, pdb.Project())
	})

	AfterEach(func() {
		//os.RemoveAll(".materialscommons")
	})

	Describe("createDirectory", func() {
		It("Should create a new directory entry", func() {
			t1DirPath := filepath.Join(projPath, "t1")
			os.MkdirAll(t1DirPath, 0777)
			fi, err := os.Stat(t1DirPath)
			Expect(err).To(BeNil())
			entry := files.TreeEntry{
				Path:  t1DirPath,
				Finfo: fi,
			}
			u.createDirectory(entry)
			d, err := u.db.FindDirectory(t1DirPath)
			Expect(err).To(BeNil())
			Expect(d.Path).To(Equal(t1DirPath))
			Expect(d.DirectoryID).To(ContainSubstring("-"))
		})
	})

	Describe("uploadFile", func() {
		It("Should upload the file", func() {
			t1DirPath := filepath.Join(projPath, "t1")
			os.MkdirAll(t1DirPath, 0777)
			fpath := filepath.Join(t1DirPath, "uploadfile.txt")
			f, err := os.Create(fpath)
			Expect(err).To(BeNil())
			f.WriteString("hello world")
			f.Close()
			d, _ := u.db.FindDirectory(t1DirPath)
			fi, _ := os.Stat(fpath)
			entry := files.TreeEntry{
				Path:  fpath,
				Finfo: fi,
			}
			u.uploadFile(entry, nil, d)
			createdFile, err := u.db.FindFile("uploadfile.txt", d.ID)
			fmt.Printf("createdFile = %#v\n", createdFile)
			Expect(err).To(BeNil())
			Expect(createdFile.FileID).To(ContainSubstring("-"))
		})
	})

	//	Describe("test marshalling", func() {
	//		It("Should properly marshal", func() {
	//			chunkSize := int32(1024*1024)
	//			uploadReq := mcstore.CreateUploadRequest{
	//				ProjectID:   "test",
	//				DirectoryID: "abc123",
	//				FileName:    "test.txt",
	//				FileSize:    11,
	//				ChunkSize:   chunkSize,
	//				FileMTime:   time.Now().Format(time.RFC1123),
	//				Checksum:    "abc123",
	//			}
	//
	//			b, err := json.Marshal(uploadReq)
	//			Expect(err).To(BeNil())
	//			fmt.Println(string(b))
	//
	//			var newReq mcstore.CreateUploadRequest
	//			err = json.Unmarshal(b, &newReq)
	//			Expect(err).To(BeNil())
	//		})
	//	})

	//	Describe("handleDirEntry", func() {
	//
	//	})
	//
	//	Describe("sendFlowReq", func() {
	//
	//	})
	//
	//	Describe("getUploadResponse", func() {
	//
	//	})
	//
	//	Describe("uploadFile", func() {
	//
	//	})
	//
	//	Describe("getDirByPath", func() {
	//
	//	})
	//
	//	Describe("getFileByName", func() {
	//
	//	})
	//
	//	Describe("handleFileEntry", func() {
	//
	//	})
	//
	//	Describe("uploadEntry", func() {
	//		It("Should", func() {
	//			Expect("").To(Equal(""))
	//		})
	//	})
})
