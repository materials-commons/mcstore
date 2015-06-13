package mc

import (
	"os"
	"time"

	"github.com/materials-commons/mcstore/pkg/app"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MCProject", func() {
	var (
		mcproject *MCProject
	)

	BeforeEach(func() {
		var err error
		os.Mkdir(".mcproject", 0777)
		project := ClientProject{
			Path:      ".mcproject",
			Name:      "proj1",
			ProjectID: "proj1id",
		}
		mcproject, err = Create(project)
		Expect(err).To(BeNil())
		Expect(mcproject).ToNot(BeNil())
	})

	AfterEach(func() {
		os.RemoveAll(".mcproject")
	})

	Describe("Create method tests", func() {
		It("Should find the created project", func() {
			db := mcproject.db
			var projects []Project
			err := db.Select(&projects, "select * from project")
			Expect(err).To(BeNil())
			Expect(projects).To(HaveLen(1))
			proj := projects[0]
			Expect(proj.ProjectID).To(Equal("proj1id"))
			Expect(proj.Name).To(Equal("proj1"))
		})
	})

	Describe("InsertDir method tests", func() {
		It("Should successfully insert and find the inserted directory", func() {
			db := mcproject.db
			now := time.Now()
			dir := &Directory{
				DirectoryID: "abc123",
				Path:        "/tmp/dir",
				LastUpload:  now,
			}

			var err error
			dir, err = mcproject.InsertDirectory(dir)
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

	Describe("FindDirectoryByPath method tests", func() {
		BeforeEach(func() {
			now := time.Now()
			dir := &Directory{
				DirectoryID: "abc123",
				Path:        "/tmp/dir",
				LastUpload:  now,
			}

			var err error
			dir, err = mcproject.InsertDirectory(dir)
			Expect(err).To(BeNil())
			Expect(dir.ID).ToNot(BeNumerically("==", 0))
		})

		It("Should find directory by path", func() {
			dir, err := mcproject.FindDirectoryByPath("/tmp/dir")
			Expect(err).To(BeNil())
			Expect(dir.DirectoryID).To(Equal("abc123"))
		})

		It("Should get ErrNotFound for directory that doesn't exist", func() {
			dir, err := mcproject.FindDirectoryByPath("/does/not/exist")
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
			dir, err = mcproject.InsertDirectory(dir)
			Expect(err).To(BeNil())
			Expect(dir.ID).ToNot(BeNumerically("==", 0))

			f := &File{
				FileID:    "fileid123",
				Directory: dir.ID,
				Name:      "test.txt",
				Size:      64 * 1024 * 1024 * 1024,
			}

			f, err = mcproject.InsertFile(f)
			Expect(err).To(BeNil())
			Expect(f.ID).NotTo(BeNumerically("==", 0))
			fid = f.ID
		})

		It("Should find inserted file", func() {
			db := mcproject.db
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
