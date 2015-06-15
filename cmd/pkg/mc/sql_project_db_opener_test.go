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

var _ = Describe("SqlProjectDbOpener", func() {
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

	Describe("OpenProjectDB method tests", func() {
		Context("ProjectDBCreate flag", func() {

			It("Should return an error if the path doesn't exist", func() {
				os.RemoveAll(".materialscommons")
				pdb, err := projectOpener.OpenProjectDB(projectDBSpec, ProjectDBCreate)
				Expect(err).To(Equal(app.ErrNotFound))
				Expect(pdb).To(BeNil())
			})

			It("Should return an error if the project already exists", func() {
				ioutil.WriteFile(filepath.Join(".materialscommons", "proj1id.db"), []byte("hello"), 0777)
				pdb, err := projectOpener.OpenProjectDB(projectDBSpec, ProjectDBCreate)
				Expect(err).To(Equal(app.ErrExists))
				Expect(pdb).To(BeNil())
			})

			It("Should create the project when it doesn't exist", func() {
				pdb, err := projectOpener.OpenProjectDB(projectDBSpec, ProjectDBCreate)
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
})
