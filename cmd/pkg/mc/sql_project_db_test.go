package mc

import (
	"time"

	"os"

	"github.com/gtarcea/1DevDayTalk2014/app"
	"github.com/materials-commons/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SQLProjectDB", func() {
	var (
		projectDB *sqlProjectDB
	)

	BeforeEach(func() {
		config.Set("mcconfigdir", ".materialscommons")
		os.Mkdir(".materialscommons", 0777)
		projectDBSpec := ProjectDBSpec{
			Path:      "/tmp",
			Name:      "proj1",
			ProjectID: "proj1id",
		}
		projectOpener := sqlProjectDBOpener{
			configer: configConfiger{},
		}
		projectDBtemp, _ := projectOpener.CreateProjectDB(projectDBSpec)
		projectDB = projectDBtemp.(*sqlProjectDB)
	})

	AfterEach(func() {
		os.RemoveAll(".materialscommons")
	})

	Describe("InsertDirectory method tests", func() {
		It("Should successfully insert and find the inserted directory", func() {
			db := projectDB.db
			now := time.Now()
			dir := &Directory{
				DirectoryID: "abc123",
				Path:        "/tmp/dir",
				LastUpload:  now,
			}

			var err error
			dir, err = projectDB.InsertDirectory(dir)
			Expect(err).To(BeNil())
			Expect(dir.ID).ToNot(BeNumerically("==", 0))

			var dirs []Directory
			err = db.Select(&dirs, "select * from directories")
			Expect(err).To(BeNil())
			Expect(dirs).To(HaveLen(1))
			d := dirs[0]
			Expect(d.DirectoryID).To(Equal("abc123"))
			Expect(d.LastUpload).To(BeTemporally("==", now))
		})
	})

	Describe("UpdateDirectory method tests", func() {
		It("Should update existing directory", func() {
			db := projectDB.db
			now := time.Now()
			dir := &Directory{
				DirectoryID: "abc123",
				Path:        "/tmp/dir",
				LastUpload:  now,
			}

			var err error
			dir, err = projectDB.InsertDirectory(dir)
			Expect(err).To(BeNil())
			Expect(dir.ID).ToNot(BeNumerically("==", 0))

			var dirs []Directory
			err = db.Select(&dirs, "select * from directories")
			Expect(err).To(BeNil())
			Expect(dirs).To(HaveLen(1))
			d := dirs[0]
			Expect(d.DirectoryID).To(Equal("abc123"))
			Expect(d.LastUpload).To(BeTemporally("==", now))

			dir.DirectoryID = "def456"
			dir.Path = "/tmp/dir2"
			err = projectDB.UpdateDirectory(dir)
			var dirs2 []Directory
			Expect(err).To(BeNil(), "Error %s", err)
			err = db.Select(&dirs2, "select * from directories")
			Expect(err).To(BeNil())
			Expect(dirs2).To(HaveLen(1))
			d = dirs2[0]
			Expect(d.DirectoryID).To(Equal("def456"))
			Expect(d.Path).To(Equal("/tmp/dir2"))
		})
	})

	Describe("FindDirectory method tests", func() {
		BeforeEach(func() {
			now := time.Now()
			dir := &Directory{
				DirectoryID: "abc123",
				Path:        "/tmp/dir",
				LastUpload:  now,
			}

			var err error
			dir, err = projectDB.InsertDirectory(dir)
			Expect(err).To(BeNil())
			Expect(dir.ID).ToNot(BeNumerically("==", 0))
		})

		It("Should find directory by path", func() {
			dir, err := projectDB.FindDirectory("/tmp/dir")
			Expect(err).To(BeNil())
			Expect(dir.DirectoryID).To(Equal("abc123"))
		})

		It("Should get ErrNotFound for directory that doesn't exist", func() {
			dir, err := projectDB.FindDirectory("/does/not/exist")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(dir).To(BeNil())
		})
	})

	Describe("InsertFile method tests", func() {
		var (
			fid   int64
			dirID int64
		)

		BeforeEach(func() {
			now := time.Now()
			dir := &Directory{
				DirectoryID: "abc123",
				Path:        "/tmp/dir",
				LastUpload:  now,
			}

			var err error
			dir, err = projectDB.InsertDirectory(dir)
			Expect(err).To(BeNil())
			Expect(dir.ID).ToNot(BeNumerically("==", 0))
			dirID = dir.ID

			f := &File{
				FileID:    "fileid123",
				Directory: dir.ID,
				Name:      "test.txt",
				Size:      64 * 1024 * 1024 * 1024,
			}

			f, err = projectDB.InsertFile(f)
			Expect(err).To(BeNil())
			Expect(f.ID).NotTo(BeNumerically("==", 0))
			fid = f.ID
		})

		It("Should find inserted file", func() {
			db := projectDB.db
			var files []File
			err := db.Select(&files, "select * from files")
			Expect(err).To(BeNil())
			Expect(files).To(HaveLen(1))
			f0 := files[0]
			Expect(f0.FileID).To(Equal("fileid123"))
			Expect(f0.Name).To(Equal("test.txt"))
			expectedSize := (64 * 1024 * 1024 * 1024)
			Expect(f0.Size).To(BeNumerically("==", expectedSize))
			Expect(f0.ID).To(Equal(fid))

			f, err := projectDB.FindFile("test.txt", dirID)
			Expect(err).To(BeNil())
			Expect(f.ID).To(BeNumerically("==", fid))
		})
	})

	Describe("UpdateFile method tests", func() {
		var fid int64
		var f *File

		BeforeEach(func() {
			now := time.Now()
			dir := &Directory{
				DirectoryID: "abc123",
				Path:        "/tmp/dir",
				LastUpload:  now,
			}

			var err error
			dir, err = projectDB.InsertDirectory(dir)
			Expect(err).To(BeNil())
			Expect(dir.ID).ToNot(BeNumerically("==", 0))

			f = &File{
				FileID:    "fileid123",
				Directory: dir.ID,
				Name:      "test.txt",
				Size:      64 * 1024 * 1024 * 1024,
			}

			f, err = projectDB.InsertFile(f)
			Expect(err).To(BeNil())
			Expect(f.ID).NotTo(BeNumerically("==", 0))
			fid = f.ID
		})

		It("Should update the file", func() {
			f.Name = "test1.txt"
			f.FileID = "fileid456"
			now := time.Now()
			f.LastDownload = now
			err := projectDB.UpdateFile(f)
			Expect(err).To(BeNil(), "Error: %s", err)

			db := projectDB.db
			var files []File
			err = db.Select(&files, "select * from files")
			Expect(err).To(BeNil())
			Expect(files).To(HaveLen(1))
			f0 := files[0]
			Expect(f0.FileID).To(Equal("fileid456"))
			Expect(f0.Name).To(Equal("test1.txt"))
			Expect(f0.LastDownload).To(BeTemporally("==", now))
		})
	})

	Describe("UpdateProject method tests", func() {
		It("Should update the project", func() {
			now := time.Now()
			proj := &Project{
				ProjectID:  "proj2id",
				LastUpload: now,
				ID:         1,
			}

			err := projectDB.UpdateProject(proj)
			Expect(err).To(BeNil())
			db := projectDB.db
			var dbproj Project
			err = db.Get(&dbproj, "select * from project")
			Expect(err).To(BeNil(), "Get failed %s", err)
			Expect(dbproj.ProjectID).To(Equal("proj2id"))
			Expect(dbproj.LastUpload).To(BeTemporally("==", now))
		})
	})

	Describe("Project", func() {
		It("Should retrieve the project", func() {
			proj := projectDB.Project()
			Expect(proj).ToNot(BeNil())
			Expect(proj.ProjectID).To(Equal("proj1id"))
		})
	})
})
