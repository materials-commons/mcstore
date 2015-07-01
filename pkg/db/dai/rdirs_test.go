package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/model"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/testdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RDirs", func() {
	var (
		rdirs   rDirs
		session *r.Session
	)

	BeforeEach(func() {
		session = testdb.RSessionMust()
		rdirs = NewRDirs(session)
	})

	Describe("Delete", func() {
		It("Should return an error if there is no matching directory id", func() {
			err := rdirs.Delete("does-not-exist")
			Expect(err).To(Equal(app.ErrNotFound))
		})

		It("Should successfully delete an existing entry", func() {
			dir := schema.NewDirectory("test123", "test@mc.org", "test", "")
			newdir, err := rdirs.Insert(&dir)
			Expect(err).To(BeNil())
			Expect(newdir).ToNot(BeNil())
			dirID := newdir.ID

			err = rdirs.Delete(dirID)
			Expect(err).To(BeNil())

			// Make sure it deleted the directory, and the project directory entry.
			rv, err := rdirs.ByID(dirID)
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(rv).To(BeNil())

			rql := model.ProjectDirs.T().GetAllByIndex("datadir_id", dirID)
			var projectDirs []schema.Project2DataDir
			err = model.ProjectDirs.Qs(session).Rows(rql, &projectDirs)
			Expect(err).To(Equal(app.ErrNotFound))
		})
	})

	Describe("ByPath", func() {
		It("Should find an existing directory", func() {
			dir, err := rdirs.ByPath("test", "test")
			Expect(err).To(BeNil())
			Expect(dir.Name).To(Equal("test"))
		})
	})
})
