package dai

import (
	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/testdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = fmt.Sprintf("")

var _ = Describe("RFiles", func() {
	var rfiles rFiles

	BeforeEach(func() {
		rfiles = NewRFiles(testdb.RSession())
	})

	Describe("ByID", func() {
		It("Should find existing file", func() {
			f, err := rfiles.ByID("testfile.txt")
			Expect(err).To(BeNil())
			Expect(f).NotTo(BeNil())
			Expect(f.ID).To(Equal("testfile.txt"))
		})

		It("Should return an error when file doesn't exist", func() {
			f, err := rfiles.ByID("does-not-exist")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(f).To(BeNil())
		})
	})

	Describe("Insert", func() {
		Context("Explicitly set ID", func() {
			It("Should properly create the file plus all the join table entries", func() {
				file := schema.NewFile("test1.txt", "test@mc.org")
				file.ID = "test1.txt" // Explicitly set the ID
				newFile, err := rfiles.Insert(&file, "test", "test")
				Expect(err).To(BeNil())
				Expect(newFile.ID).To(Equal("test1.txt"))

				// Check that the join tables were updated.
				session := testdb.RSession()
				var p2df []schema.Project2DataFile
				rql := r.Table("project2datafile").Filter(r.Row.Field("datafile_id").Eq(file.ID))
				err = model.ProjectFiles.Qs(session).Rows(rql, &p2df)
				Expect(err).To(BeNil())
				Expect(len(p2df)).To(BeNumerically("==", 1))

				var dir2df []schema.DataDir2DataFile
				rql = r.Table("datadir2datafile").Filter(r.Row.Field("datafile_id").Eq(file.ID))
				err = model.DirFiles.Qs(session).Rows(rql, &dir2df)
				Expect(err).To(BeNil())
				Expect(len(dir2df)).To(BeNumerically("==", 1))

				deleteFile(file.ID)
			})

			It("Should return an error when attempting to insert same object twice", func() {
				tfile := schema.NewFile("test1.txt", "test@mc.org")
				tfile.ID = "test1.txt"
				newFile, err := rfiles.Insert(&tfile, "test", "test")
				Expect(err).To(BeNil())
				Expect(newFile.ID).To(Equal("test1.txt"))

				newFile, err = rfiles.Insert(&tfile, "test", "test")
				deleteFile("test1.txt")
			})
		})

		Context("Database sets ID", func() {
			It("Should properly create the file plus all the join table entries", func() {
				file := schema.NewFile("test1.txt", "test@mc.org")
				newFile, err := rfiles.Insert(&file, "test", "test")
				Expect(err).To(BeNil())
				Expect(newFile.ID).NotTo(Equal(""))

				// Check that the join tables were updated.
				session := testdb.RSession()
				var p2df []schema.Project2DataFile
				rql := r.Table("project2datafile").Filter(r.Row.Field("datafile_id").Eq(newFile.ID))
				err = model.ProjectFiles.Qs(session).Rows(rql, &p2df)
				Expect(err).To(BeNil())
				Expect(len(p2df)).To(BeNumerically("==", 1))

				var dir2df []schema.DataDir2DataFile
				rql = r.Table("datadir2datafile").Filter(r.Row.Field("datafile_id").Eq(newFile.ID))
				err = model.DirFiles.Qs(session).Rows(rql, &dir2df)
				Expect(err).To(BeNil())
				Expect(len(dir2df)).To(BeNumerically("==", 1))

				deleteFile(newFile.ID)
			})
		})
	})

	Describe("Delete", func() {
		Context("Test against existing file", func() {
			BeforeEach(func() {
				tfile := schema.NewFile("tfile", "test@mc.org")
				tfile.ID = "tfile"
				rfiles.Insert(&tfile, "test", "test")
			})

			AfterEach(func() {
				model.Files.Qs(rfiles.session).Delete("tfile")

				rql := model.DirFiles.T().GetAllByIndex("datafile_id", "tfile").
					Filter(r.Row.Field("datadir_id").Eq("test")).Delete()
				rql.RunWrite(rfiles.session)

				rql = model.ProjectFiles.T().GetAllByIndex("datafile_id", "tfile").
					Filter(r.Row.Field("project_id").Eq("test")).Delete()
				rql.RunWrite(rfiles.session)
			})

			Context("Supporting Method Tests", func() {
				Describe("getProjects", func() {
					It("Should find tfile in the test project", func() {
						projects, err := rfiles.getProjects("tfile")
						Expect(err).To(BeNil())
						Expect(len(projects)).To(BeNumerically("==", 1))
						proj := projects[0]
						Expect(proj.DataFileID).To(Equal("tfile"))
						Expect(proj.ProjectID).To(Equal("test"))
					})
				})

				Describe("getDirs", func() {
					It("Should find tfile in the test dir", func() {
						dirs, err := rfiles.getDirs("tfile")
						Expect(err).To(BeNil())
						Expect(len(dirs)).To(BeNumerically("==", 1))
						dir := dirs[0]
						Expect(dir.DataFileID).To(Equal("tfile"))
						Expect(dir.DataDirID).To(Equal("test"))
					})
				})

				Describe("deleteFromDir", func() {
					It("Should delete the tfile entry in datadir2datafile table", func() {
						err := rfiles.deleteFromDir("tfile", "test")
						Expect(err).To(BeNil())
						dirs, err := rfiles.getDirs("tfile")
						Expect(err).To(Equal(app.ErrNotFound))
						Expect(dirs).To(BeNil())
					})
				})

				Describe("deleteFromProject", func() {
					It("Should delete the tfile entry in project2datafile table", func() {
						err := rfiles.deleteFromProject("tfile", "test")
						Expect(err).To(BeNil())
						projects, err := rfiles.getProjects("tfile")
						Expect(err).To(Equal(app.ErrNotFound))
						Expect(projects).To(BeNil())
					})
				})
			})

			Context("Use Delete method to delete", func() {
				It("Should delete tfile and all the join table entries", func() {
					_, err := rfiles.Delete("tfile", "test", "test")
					Expect(err).To(BeNil())

					file, err := rfiles.ByID("tfile")
					Expect(err).To(Equal(app.ErrNotFound))
					Expect(file).To(BeNil())

					dirs, err := rfiles.getDirs("tfile")
					Expect(err).To(Equal(app.ErrNotFound))
					Expect(dirs).To(BeNil())

					projects, err := rfiles.getProjects("tfile")
					Expect(err).To(Equal(app.ErrNotFound))
					Expect(projects).To(BeNil())
				})
			})
		})
	})
})

func deleteFile(fileID string) {
	session := testdb.RSession()
	model.Files.Qs(session).Delete(fileID)

	rql := r.Table("project2datafile").Filter(r.Row.Field("datafile_id").Eq(fileID))
	var p2df []schema.Project2DataFile
	model.ProjectFiles.Qs(session).Rows(rql, &p2df)
	for _, entry := range p2df {
		model.ProjectFiles.Qs(session).Delete(entry.ID)
	}

	rql = r.Table("datadir2datafile").Filter(r.Row.Field("datafile_id").Eq(fileID))
	var d2df []schema.DataDir2DataFile
	model.DirFiles.Qs(session).Rows(rql, &d2df)
	for _, entry := range d2df {
		model.DirFiles.Qs(session).Delete(entry.ID)
	}
}
