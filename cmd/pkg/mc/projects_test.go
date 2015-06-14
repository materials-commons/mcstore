package mc

import (
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("Projects", func() {
	var (
		configer Configer
		projects *mcprojects
	)

	BeforeEach(func() {
		configer = configConfiger{}
		config.Set("mcconfigdir", ".materialscommons")
		projects = NewProjects(configer)
	})

	AfterEach(func() {
		os.RemoveAll(".materialscommons")
	})

	Describe("All method tests", func() {
		It("Should return an error when there is no configuration directory", func() {
			projs, err := projects.All()
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(projs).To(BeNil())
		})

		It("Should return an empty list when there is an empty configuration directory", func() {
			err := os.MkdirAll(".materialscommons", 0777)
			Expect(err).To(BeNil())
			projs, err := projects.All()
			Expect(err).To(BeNil())
			Expect(projs).To(HaveLen(0))
		})
	})

	Describe("Create method tests", func() {
		It("Should return an error when creating an existing project", func() {

		})

		It("Should return an error when there is no configuration directory", func() {

		})

		It("Should succeed when creating a project", func() {

		})

		It("Should add the new created project to the list of projects", func() {

		})
	})
})
