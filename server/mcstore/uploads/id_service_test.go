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
			upload = nil
		}
	})

	Describe("ID", func() {
		Describe("Access permissions", func() {
			var (
				req IDRequest
			)

			BeforeEach(func() {
				req = IDRequest{
					ProjectID:   "test",
					DirectoryID: "test",
					Host:        "host",
				}
			})

			Context("Access allowed", func() {
				It("Should allow access admin user", func() {
					req.User = "admin@mc.org"
					upload, err := s.ID(req)
					Expect(err).To(BeNil(), "Unexpected error: %s", err)
					Expect(upload).NotTo(BeNil(), "upload is nil")
				})

				It("Should allow access to user in project", func() {
					req.User = "test1@mc.org"
					upload, err := s.ID(req)
					Expect(err).To(BeNil(), "Unexpected error: %s", err)
					Expect(upload).NotTo(BeNil())
				})
			})

			Context("Access not allowed", func() {
				It("Should not allow access for users not in project", func() {
					req.User = "test2@mc.org"
					upload, err := s.ID(req)
					Expect(err).NotTo(BeNil())
					Expect(err).To(Equal(app.ErrNoAccess))
					Expect(upload).To(BeNil())
				})
			})

			Context("Invalid user", func() {
				It("Should not allow access for non existent directory", func() {
					req.User = "test@mc.org" // valid user
					req.DirectoryID = "test@mc.org"
					upload, err := s.ID(req)
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

	Describe("Delete", func() {
		Context("Access Permissions", func() {
			var (
				req IDRequest
				u   *schema.Upload
			)

			BeforeEach(func() {
				req = IDRequest{
					ProjectID:   "test",
					DirectoryID: "test",
					Host:        "host",
					User:        "test@mc.org",
				}

				u, _ = s.ID(req)
			})

			AfterEach(func() {
				if u != nil {
					s.Delete(u.ID, req.User)
					u = nil
				}
			})

			It("Should fail on user not in project", func() {
				err := s.Delete(u.ID, "test2@mc.org")
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(app.ErrNoAccess))
			})

			It("Should succeed on user in project", func() {
				err := s.Delete(u.ID, "test@mc.org")
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
			})

			It("Should succeed on admin user", func() {
				err := s.Delete(u.ID, "admin@mc.org")
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
			})

			It("Should fail on non-existant user", func() {
				err := s.Delete(u.ID, "no-such-user@doesnot.exist.com")
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(app.ErrNoAccess))
			})
		})

		Context("request ID", func() {
			It("Should fail on bad id", func() {
				err := s.Delete("no-such-id", "admin@mc.org")
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(app.ErrNotFound))
			})

			It("Should succeed on good id", func() {
				req := IDRequest{
					ProjectID:   "test",
					DirectoryID: "test",
					Host:        "host",
					User:        "admin@mc.org",
				}

				upload, err := s.ID(req)
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
				Expect(upload).NotTo(BeNil())

				err = s.Delete(upload.ID, "admin@mc.org")
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
			})
		})
	})

	PDescribe("ListForProject", func() {
		PContext("Access Permissions", func() {

		})
	})
})
