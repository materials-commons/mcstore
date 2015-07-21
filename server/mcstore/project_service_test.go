package mcstore

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai/mocks"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	amocks "github.com/materials-commons/mcstore/pkg/domain/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProjectService", func() {
	var (
		mprojects *mocks.Projects2
		mdirs     *mocks.Dirs2
		maccess   *amocks.Access
		s         *projectService
		p         *schema.Project = &schema.Project{
			ID:    "proj1",
			Name:  "proj1",
			Owner: "a@b.com",
		}
	)

	BeforeEach(func() {
		mprojects = mocks.NewMProjects2()
		mdirs = mocks.NewMDirs2()
		maccess = amocks.NewMAccess()
		s = &projectService{
			projects: mprojects,
			dirs:     mdirs,
			access:   maccess,
		}
	})

	Describe("createProject", func() {
		It("Should succeed if project exists, and mustNotExist is false", func() {
			mprojects.On("ByName").SetError(nil).SetProject(p)
			proj, exists, err := s.createProject("proj1", "a@b.com", false)
			Expect(err).To(BeNil())
			Expect(exists).To(BeTrue())
			Expect(proj.ID).To(Equal("proj1"))
		})

		It("Should fail if project exists and mustNotExist is true", func() {
			mprojects.On("ByName").SetError(nil).SetProject(p)
			proj, exists, err := s.createProject("proj1", "a@b.com", true)
			Expect(err).To(Equal(app.ErrExists))
			Expect(exists).To(BeTrue())
			Expect(proj).To(BeNil())
		})

		It("Should succeed if project doesn't exist", func() {
			mprojects.On("ByName").SetError(app.ErrNotFound).SetProject(nil)
			mprojects.On("Insert").SetError(nil).SetProject(p)
			proj, exists, err := s.createProject("proj1", "a@b.com", true)
			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
			Expect(proj.Name).To(Equal("proj1"))
			Expect(proj.ID).To(Equal("proj1"))
		})
	})

	Describe("getProjectByName", func() {
		It("Should fail if project doesn't exist", func() {
			mprojects.On("ByName").SetError(app.ErrNotFound).SetProject(nil)
			proj, err := s.getProjectByName("doesn't-exist", "a@b.com", "a@b.com")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(proj).To(BeNil())
		})

		It("Should fail if project exists but user doesn't have access", func() {
			mprojects.On("ByName").SetError(nil).SetProject(p)
			maccess.On("AllowedByOwner", "proj1", "b@c.com").Return(false)
			proj, err := s.getProjectByName("proj1", "a@b.com", "b@c.com")
			Expect(err).To(Equal(app.ErrNoAccess))
			Expect(proj).To(BeNil())
		})

		It("Should succeed if project exists and user has access", func() {
			mprojects.On("ByName").SetError(nil).SetProject(p)
			maccess.On("AllowedByOwner", "proj1", "b@c.com").Return(true)
			proj, err := s.getProjectByName("proj1", "a@b.com", "b@c.com")
			Expect(err).To(BeNil())
			Expect(proj.Name).To(Equal("proj1"))
			Expect(proj.ID).To(Equal("proj1"))
		})
	})

	Describe("getProjectByID", func() {
		It("Should fail if project doesn't exist", func() {
			mprojects.On("ByID").SetError(app.ErrNotFound).SetProject(nil)
			proj, err := s.getProjectByID("doesn't-exist", "a@b.com")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(proj).To(BeNil())
		})

		It("Should fail if project exists but user doesn't have access", func() {
			mprojects.On("ByID").SetError(nil).SetProject(p)
			maccess.On("AllowedByOwner", "proj1", "b@c.com").Return(false)
			proj, err := s.getProjectByID("proj1", "b@c.com")
			Expect(err).To(Equal(app.ErrNoAccess))
			Expect(proj).To(BeNil())
		})

		It("Should succeed if project exists and user has access", func() {
			mprojects.On("ByID").SetError(nil).SetProject(p)
			maccess.On("AllowedByOwner", "proj1", "b@c.com").Return(true)
			proj, err := s.getProjectByID("proj1", "b@c.com")
			Expect(err).To(BeNil())
			Expect(proj.Name).To(Equal("proj1"))
			Expect(proj.ID).To(Equal("proj1"))
		})
	})
})
