package mc

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SqlProjectDBOpener", func() {
	Context("Open/Create Tests", func() {
		var (
			projectDBSpec ProjectDBSpec
			projectOpener sqlProjectDBOpener = sqlProjectDBOpener{
				configer: configConfiger{},
			}
		)
		BeforeEach(func() {
			config.Set("mcconfigdir", ".materialscommons")
			os.Mkdir(".materialscommons", 0777)
			projectDBSpec = ProjectDBSpec{
				Path:      "/tmp",
				Name:      "proj1",
				ProjectID: "proj1id",
			}
		})

		AfterEach(func() {
			os.RemoveAll(".materialscommons")
		})

		Describe("CreateProjectDB method tests", func() {
			It("Should return an error if the path doesn't exist", func() {
				os.RemoveAll(".materialscommons")
				pdb, err := projectOpener.CreateProjectDB(projectDBSpec)
				Expect(err).To(Equal(app.ErrNotFound))
				Expect(pdb).To(BeNil())
			})

			It("Should return an error if the project already exists", func() {
				ioutil.WriteFile(filepath.Join(".materialscommons", "proj1.db"), []byte("hello"), 0777)
				pdb, err := projectOpener.CreateProjectDB(projectDBSpec)
				Expect(err).To(Equal(app.ErrExists))
				Expect(pdb).To(BeNil())
			})

			It("Should create the project when it doesn't exist", func() {
				pdb, err := projectOpener.CreateProjectDB(projectDBSpec)
				Expect(err).To(BeNil())
				Expect(pdb).NotTo(BeNil())

				// Test the projects contents
				sqlpdb := pdb.(*sqlProjectDB)
				db := sqlpdb.db
				var projects []Project
				err = db.Select(&projects, "select * from project")
				Expect(err).To(BeNil())
				Expect(projects).To(HaveLen(1))
				proj := projects[0]
				Expect(proj.ProjectID).To(Equal("proj1id"))
				Expect(proj.Name).To(Equal("proj1"))
				Expect(proj.Path).To(Equal("/tmp"))
			})
		})

		Describe("OpenProjectDB method tests", func() {
			It("Should return an error when the project doesn't exist", func() {
				pdb, err := projectOpener.OpenProjectDB("does-not-exist")
				Expect(err).To(Equal(app.ErrNotFound))
				Expect(pdb).To(BeNil())
			})

			It("Should open an existing project", func() {
				// Create the project
				pdb, err := projectOpener.CreateProjectDB(projectDBSpec)
				Expect(err).To(BeNil())
				Expect(pdb).NotTo(BeNil())

				// Open the project and test its contents
				pdb, err = projectOpener.OpenProjectDB("proj1")
				Expect(err).To(BeNil())
				Expect(pdb).NotTo(BeNil())
				sqlpdb := pdb.(*sqlProjectDB)
				db := sqlpdb.db
				var projects []Project
				err = db.Select(&projects, "select * from project")
				Expect(err).To(BeNil())
				Expect(projects).To(HaveLen(1))
				proj := projects[0]
				Expect(proj.ProjectID).To(Equal("proj1id"))
				Expect(proj.Name).To(Equal("proj1"))
			})
		})
	})

	Describe("PathToName method tests", func() {
		var (
			projectOpener sqlProjectDBOpener = sqlProjectDBOpener{
				configer: configConfiger{},
			}
		)

		It("Should return last element of path that has .db extension without .db", func() {
			path := "/tmp/this.db"
			name := projectOpener.PathToName(path)
			Expect(name).To(Equal("this"))
		})

		It("Should return return last element of path for a name that has multiple dots without .db extension", func() {
			path := "/tmp/this.is.name.db"
			name := projectOpener.PathToName(path)
			Expect(name).To(Equal("this.is.name"))
		})
	})
})
