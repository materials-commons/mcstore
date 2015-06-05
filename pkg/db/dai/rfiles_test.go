package dai

import (
	"fmt"

	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RFiles", func() {
	var rfiles Files

	BeforeEach(func() {
		rfiles = NewRFiles(testutil.RSession())
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
				session := testutil.RSession()
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
		})

		Context("Database sets ID", func() {
			It("Should properly create the file plus all the join table entries", func() {
				file := schema.NewFile("test1.txt", "test@mc.org")

				fmt.Printf("%#v", rfiles)
				newFile, err := rfiles.Insert(&file, "test", "test")
				Expect(err).To(BeNil())
				Expect(newFile.ID).NotTo(Equal(""))

				// Check that the join tables were updated.
				session := testutil.RSession()
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
})

func deleteFile(fileID string) {
	session := testutil.RSession()
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
