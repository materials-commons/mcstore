package dai

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/testdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RUploads", func() {
	var ruploads Uploads

	BeforeEach(func() {
		ruploads = NewRUploads(testdb.RSession())
	})

	Describe("For User", func() {
		It("Should return an error if user can't be found", func() {
			uploads, err := ruploads.ForUser("no-such-user")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(uploads).To(BeNil())
		})

		It("Should find uploads for user", func() {
			upload := schema.CUpload().
				Owner("test@mc.org").
				Project("test", "test").
				Directory("test", "test").
				Create()
			newUpload, err := ruploads.Insert(&upload)
			Expect(err).To(BeNil())
			Expect(newUpload.Owner).To(Equal("test@mc.org"))
			uploads, err := ruploads.ForUser("test@mc.org")
			Expect(err).To(BeNil())
			Expect(uploads).To(HaveLen(1))
			err = ruploads.Delete(uploads[0].ID)
			Expect(err).To(BeNil())
		})
	})

	Describe("Update", func() {
		It("Should update fields for an upload", func() {
			upload := schema.CUpload().
				Owner("test@mc.org").
				Project("test", "test").
				Create()
			newUpload, err := ruploads.Insert(&upload)
			Expect(err).To(BeNil())
			Expect(newUpload.Owner).To(Equal("test@mc.org"))

			newUpload.ProjectName = "changedName"
			err = ruploads.Update(newUpload)
			Expect(err).To(BeNil())

			newUpload, err = ruploads.ByID(newUpload.ID)
			Expect(err).To(BeNil())
			Expect(newUpload.ProjectName).To(Equal("changedName"))
			err = ruploads.Delete(newUpload.ID)
			Expect(err).To(BeNil())
		})
	})
})
