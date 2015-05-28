package mcstore

import (
	dmocks "github.com/materials-commons/mcstore/pkg/db/dai/mocks"
	amocks "github.com/materials-commons/mcstore/pkg/domain/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DirService", func() {
	Describe("createDir Method Tests", func() {
		var (
			mdirs      *dmocks.Dirs
			mprojects  *dmocks.Projects
			maccess    *amocks.Access
			nilProject *schema.Project
			nilDir     *schema.Directory
			s          *dirService
		)

		BeforeEach(func() {
			mdirs = dmocks.NewMDirs()
			mprojects = dmocks.NewMProjects()
			maccess = amocks.NewMAccess()
			s = &dirService{
				dirs:     mdirs,
				projects: mprojects,
				access:   maccess,
			}
		})

		It("Should fail when project doesn't exist", func() {
			mprojects.On("ByID", "does-not-exist").Return(nilProject, app.ErrNotFound)
			dir, err := s.createDir("does-not-exist", "/some/path")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(dir).To(BeNil())
		})

		It("Should fail when project exists but path is invalid", func() {
			proj := &schema.Project{
				Name: "proj1",
				ID:   "proj1",
			}
			mprojects.On("ByID", "proj1").Return(proj, nil)
			dir, err := s.createDir("proj1", "bad/path")
			Expect(err).To(Equal(app.ErrInvalid))
			Expect(dir).To(BeNil())
		})

		It("Should succeed when project and path are valid and directory doesn't exist", func() {
			proj := &schema.Project{
				Name: "proj1",
				ID:   "proj1",
			}
			d := &schema.Directory{
				ID:   "dir1",
				Name: "dir1",
			}

			mprojects.On("ByID", "proj1").Return(proj, nil)
			mdirs.On("Insert", mock.AnythingOfType("*schema.Directory")).Return(d, nil)
			mdirs.On("ByPath", "proj1/dir1", "proj1").Return(nilDir, app.ErrNotFound)
			dir, err := s.createDir("proj1", "proj1/dir1")
			Expect(err).To(BeNil())
			Expect(dir.Name).To(Equal("dir1"))
			Expect(dir.ID).To(Equal("dir1"))
		})

	})

	Describe("validDirPath Method Tests", func() {

	})
})
