package uploads

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/materials-commons/mcstore/pkg/domain"
	"github.com/materials-commons/mcstore/testdb"
)

var _ = fmt.Println

var _ = Describe("IDService", func() {

	var (
		users    = dai.NewRUsers(testdb.RSession())
		files    = dai.NewRFiles(testdb.RSession())
		dirs     = dai.NewRDirs(testdb.RSession())
		projects = dai.NewRProjects(testdb.RSession())
		uploads  = dai.NewRUploads(testdb.RSession())
		access   = domain.NewAccess(projects, files, users)
		s        = &idService{
			files:       files,
			dirs:        dirs,
			projects:    projects,
			uploads:     uploads,
			access:      access,
			fops:        file.MockOps(),
			tracker:     requestBlockTracker,
			requestPath: &mockRequestPath{},
		}
		upload *schema.Upload
	)

	BeforeEach(func() {
		s.fops = file.MockOps()
	})

	AfterEach(func() {
		if upload != nil {
			err := uploads.Delete(upload.ID)
			Expect(err).To(BeNil())
			upload = nil
		}
	})

	Describe("ID Method Tests", func() {
		Describe("Access permissions", func() {
			var (
				req IDRequest
			)

			BeforeEach(func() {
				req = IDRequest{
					ProjectID:   "test",
					DirectoryID: "test",
					Host:        "host",
					FileSize:    10,
					ChunkSize:   10,
				}
			})

			Context("Access allowed", func() {
				It("Should allow access to admin user", func() {
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

			Context("Invalid directory", func() {
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
						FileSize:    10,
						ChunkSize:   10,
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

		Describe("Existing Uploads", func() {
			It("Should find an existing upload", func() {
				req := IDRequest{
					ProjectID:   "test",
					DirectoryID: "test",
					Host:        "host",
					Checksum:    "abc123",
					FileSize:    100,
					ChunkSize:   10,
					User:        "test@mc.org",
				}
				upload, err := s.ID(req)
				Expect(err).To(BeNil())
				Expect(upload.File.Blocks.Len()).To(BeNumerically("==", 10))

				// Now submit again with exact same parameters. It should
				// not create a new request
				upload2, err := s.ID(req)
				Expect(err).To(BeNil())
				Expect(upload2.ID).To(Equal(upload.ID))

				// Check that the bitset state is correct.
				Expect(upload2.File.Blocks.Len()).To(BeNumerically("==", 10))
			})

			It("Should leave an upload in place", func() {
				req := IDRequest{
					ProjectID:   "test",
					DirectoryID: "test",
					Host:        "host",
					Checksum:    "abc123567",
					FileSize:    10,
					ChunkSize:   10,
					User:        "test@mc.org",
					FileName:    "blocktest.txt",
				}
				s.ID(req)
			})
		})
	})

	Describe("Delete Method Tests", func() {
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
					FileSize:    10,
					ChunkSize:   10,
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

		Context("request ID Validation", func() {
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
					FileSize:    10,
					ChunkSize:   10,
				}

				upload, err := s.ID(req)
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
				Expect(upload).NotTo(BeNil())

				err = s.Delete(upload.ID, "admin@mc.org")
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
			})
		})
	})

	Describe("UploadsForProject Method Tests", func() {
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
				FileSize:    10,
				ChunkSize:   10,
			}

			u, _ = s.ID(req)
		})

		AfterEach(func() {
			if u != nil {
				s.Delete(u.ID, req.User)
				u = nil
			}
		})

		Context("Access Permissions", func() {
			It("Should fail on user not in project", func() {
				uploads, err := s.UploadsForProject("test", "test2@mc.org")
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(app.ErrNoAccess))
				Expect(uploads).To(BeNil())
			})

			It("Should succeed on user in project", func() {
				uploads, err := s.UploadsForProject("test", "test@mc.org")
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
				Expect(len(uploads)).To(BeNumerically(">", 0))
			})

			It("Should succeed on admin user", func() {
				uploads, err := s.UploadsForProject("test", "admin@mc.org")
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
				Expect(len(uploads)).To(BeNumerically(">", 0))
			})

			It("Should fail on non-existent user", func() {
				uploads, err := s.UploadsForProject("test", "no-such-user@doesnot.exist.com")
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(app.ErrNoAccess))
				Expect(uploads).To(BeNil())
			})
		})

		Context("Project ID Validation", func() {
			It("Should fail on bad project", func() {
				uploads, err := s.UploadsForProject("no-such-project", "test@mc.org")
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(app.ErrInvalid))
				Expect(uploads).To(BeNil())
			})

			It("Should succeed on good project", func() {
				uploads, err := s.UploadsForProject("test", "test@mc.org")
				Expect(err).To(BeNil(), "Unexpected error: %s", err)
				Expect(len(uploads)).To(BeNumerically(">", 0))
			})
		})
	})
})
