package uploads

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/materials-commons/mcstore/test"
)

var _ = Describe("IDService", func() {

	var (
		users    = dai.NewRUsers(test.RSession())
		files    = dai.NewRFiles(test.RSession())
		dirs     = dai.NewRDirs(test.RSession())
		projects = dai.NewRProjects(test.RSession())
		uploads  = dai.NewRUploads(test.RSession())
		access   = domain.NewAccess(projects, files, users)
		s        = NewIDServiceFrom(dirs, projects, uploads, access)
		upload   *schema.Upload
	)

	AfterEach(func() {
		if upload != nil {
			err := uploads.Delete(upload.ID)
			Expect(err).To(BeNil())
		}
	})

	Describe("ID", func() {
		Describe("Access permissions", func() {
			var (
				cf IDRequest
			)

			BeforeEach(func() {
				cf = IDRequest{
					ProjectID:   "test",
					DirectoryID: "test",
					Host:        "host",
				}
			})

			Context("Access allowed", func() {
				It("Should allow access admin user", func() {
					cf.User = "admin@mc.org"
					upload, err := s.ID(cf)
					Expect(err).To(BeNil(), "Unexpected error: %s", err)
					Expect(upload).NotTo(BeNil(), "upload is nil")
				})

				It("Should allow access to user in project", func() {
					cf.User = "test1@mc.org"
					upload, err := s.ID(cf)
					Expect(err).To(BeNil(), "Unexpected error: %s", err)
					Expect(upload).NotTo(BeNil())
				})
			})

			Context("Access not allowed", func() {
				It("Should not allow access for users not in project", func() {
					cf.User = "test2@mc.org"
					upload, err := s.ID(cf)
					Expect(err).NotTo(BeNil())
					Expect(err).To(Equal(app.ErrNoAccess))
					Expect(upload).To(BeNil())
				})
			})

			Context("Invalid user", func() {
				It("Should not allow access for non existent directory", func() {
					cf.User = "test@mc.org" // valid user
					cf.DirectoryID = "test@mc.org"
					upload, err := s.ID(cf)
					Expect(err).NotTo(BeNil())
					Expect(upload).To(BeNil())
				})
			})
		})

		Describe("Request Parameters", func() {
			Context("Bad Request", func() {
				var req IDRequest

				BeforeEach(func() {
					req = IDRequest{
						User:        "admin@mc.org",
						ProjectID:   "test",
						DirectoryID: "test",
					}
				})

				It("Should fail on bad project id", func() {
					req.ProjectID = "does-not-exist"
					upload, err := s.ID(req)
					Expect(err).To(HaveOccurred())
					Expect(upload).To(BeNil())
				})

				It("Should fail on bad directory id", func() {
					req.DirectoryID = "does-not-exist"
					upload, err := s.ID(req)
					Expect(err).To(HaveOccurred())
					Expect(upload).To(BeNil())
				})

				It("Should fail on directory id not in project", func() {
					req.DirectoryID = "test2" // in different project (test2)
					upload, err := s.ID(req)
					Expect(err).To(HaveOccurred())
					Expect(upload).To(BeNil())
				})
			})
		})
	})
})
