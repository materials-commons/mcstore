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

	Describe("InsertDir method tests", func() {
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
		var fid int64

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
		})
	})
})
