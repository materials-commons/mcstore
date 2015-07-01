package mc

import (
	"os"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Projects", func() {
	var (
		configer Configer
		projects *mcprojects
	)

	BeforeEach(func() {
		os.MkdirAll(".materialscommons", 0777)
		configer = configConfiger{}
		config.Set("mcconfigdir", ".materialscommons")
		projects = NewProjects(configer)
	})

	AfterEach(func() {
		os.RemoveAll(".materialscommons")
	})

	Describe("All method tests", func() {
		It("Should return an error when there is no configuration directory", func() {
			os.RemoveAll(".materialscommons")
			projs, err := projects.All()
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(projs).To(BeNil())
		})

		It("Should return an empty list when there is an empty configuration directory", func() {
			projs, err := projects.All()
			Expect(err).To(BeNil())
			Expect(projs).To(HaveLen(0))
		})

		It("Should return the list of projects when there is a project", func() {
			projectOpener := sqlProjectDBOpener{
				configer: configer,
			}

			projectDBSpec := ProjectDBSpec{
				Path:      "/tmp",
				Name:      "proj1",
				ProjectID: "proj1id",
			}
			_, err := projectOpener.CreateProjectDB(projectDBSpec)
			Expect(err).To(BeNil())

			projs, err := projects.All()
			Expect(err).To(BeNil())
			Expect(projs).To(HaveLen(1))
		})
	})

	Describe("Create method tests", func() {
		var (
			projectDBSpec ProjectDBSpec
		)

		BeforeEach(func() {
			os.MkdirAll(".materialscommons", 0777)
			projectDBSpec = ProjectDBSpec{
				Path:      "/tmp",
				Name:      "proj1",
				ProjectID: "proj1id",
			}
		})

		AfterEach(func() {
			os.RemoveAll(".materialscommons")
		})

		It("Should succeed when creating a project", func() {
			pdb, err := projects.Create(projectDBSpec)
			Expect(err).To(BeNil())
			Expect(pdb).NotTo(BeNil())
		})

		It("Should return an error when creating an existing project", func() {
			pdb, err := projects.Create(projectDBSpec)
			Expect(err).To(BeNil())
			Expect(pdb).NotTo(BeNil())

			pdb, err = projects.Create(projectDBSpec)
			Expect(err).To(Equal(app.ErrExists))
			Expect(pdb).To(BeNil())
		})

		It("Should return an error when there is no configuration directory", func() {
			os.RemoveAll(".materialscommons")
			pdb, err := projects.Create(projectDBSpec)
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(pdb).To(BeNil())
		})

		It("Should add the new created project to the list of projects", func() {
			projs, err := projects.All()
			Expect(err).To(BeNil())
			Expect(projs).To(HaveLen(0))

			pdb, err := projects.Create(projectDBSpec)
			Expect(err).To(BeNil())
			Expect(pdb).NotTo(BeNil())

			projs, err = projects.All()
			Expect(err).To(BeNil())
			Expect(projs).To(HaveLen(1))
		})
	})
})
