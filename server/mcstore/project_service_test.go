package mcstore

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/dai/mocks"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProjectService", func() {
	var (
		mprojects *mocks.Projects2
		mdirs     *mocks.Dirs2
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
		s = &projectService{
			projects: mprojects,
			dirs:     mdirs,
		}
	})

	Describe("createProject Method Tests", func() {
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
			mprojects.On("ByName").SetError(nil).SetProject(nil)
			mprojects.On("Insert").SetError(nil).SetProject(p)
			proj, exists, err := s.createProject("proj1", "a@b.com", true)
			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
			Expect(proj.Name).To(Equal("proj1"))
			Expect(proj.ID).To(Equal("proj1"))
		})
	})

	//	Describe("getProject Method Tests", func() {
	//		It("Should succeed if project exists and create is false", func() {
	//			mprojects.On("ByName").SetError(nil).SetProject(p)
	//			proj, created, err := s.getProject("proj1", "a@b.com", false)
	//			Expect(err).To(BeNil())
	//			Expect(created).To(BeFalse())
	//			Expect(proj).NotTo(BeNil())
	//		})
	//
	//		It("Should succeed if project doesn't exist and create is true", func() {
	//			mprojects.On("ByName").SetError(app.ErrNotFound).SetProject(nil)
	//			mprojects.On("Insert").SetError(nil).SetProject(p)
	//			proj, created, err := s.getProject("proj1", "a@b.com", true)
	//			Expect(err).To(BeNil())
	//			Expect(created).To(BeTrue())
	//			Expect(proj).NotTo(BeNil())
	//		})
	//
	//		It("Should fail if project doesn't exist and create is false", func() {
	//			mprojects.On("ByName").SetError(app.ErrNotFound).SetProject(nil)
	//			proj, created, err := s.getProject("proj1", "a@b.com", false)
	//			Expect(err).To(Equal(app.ErrNotFound))
	//			Expect(created).To(BeFalse())
	//			Expect(proj).To(BeNil())
	//		})
	//	})
})
