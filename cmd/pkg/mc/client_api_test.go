package mc

import (
	"fmt"
	"net/http/httptest"
	"os"

	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/testdb"
	"github.com/materials-commons/mcstore/server/mcstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = fmt.Println

var _ = Describe("ClientApi", func() {
	var (
		projectDBSpec ProjectDBSpec
		projectOpener sqlProjectDBOpener = sqlProjectDBOpener{
			configer: configConfiger{},
		}
		projectDB ProjectDB
		server    *httptest.Server
	)

	BeforeEach(func() {
		config.Set("mcconfigdir", ".materialscommons")
		os.Mkdir(".materialscommons", 0777)
		projectDBSpec = ProjectDBSpec{
			Path:      "/tmp",
			Name:      "test2",
			ProjectID: "test2",
		}
		projectDB, _ = projectOpener.CreateProjectDB(projectDBSpec)

		container := mcstore.NewServicesContainer(testdb.Sessions)
		server = httptest.NewServer(container)
		config.Set("mcurl", server.URL)
		config.Set("apikey", "test")
	})

	AfterEach(func() {
		os.RemoveAll(".materialscommons")
	})

	Describe("CreateDirectory", func() {
		It("Should fail on a bad project", func() {
			c := newClientAPIWithConfiger(configConfiger{})
			dirID, err := c.CreateDirectory("does-not-exist", "/does/not/matter")
			Expect(err).To(MatchError(app.ErrNotFound))
			Expect(dirID).To(Equal(""))
		})

		It("Should fail on a bad directory name", func() {
			c := newClientAPIWithConfiger(configConfiger{})
			dirID, err := c.CreateDirectory("test2", "/this/that")
			Expect(err).To(MatchError(app.ErrInvalid))
			Expect(dirID).To(Equal(""))
		})

		It("Should succeed with windows style path", func() {
			c := newClientAPIWithConfiger(configConfiger{})
			dirID, err := c.CreateDirectory("test2", `c:\tmp\test2\mydir`)
			Expect(err).To(Succeed())
			Expect(dirID).NotTo(Equal(""))
		})

		It("Should succeed with Linux style path", func() {
			c := newClientAPIWithConfiger(configConfiger{})
			dirID, err := c.CreateDirectory("test2", "/tmp/test2/mydir2")
			Expect(err).To(Succeed())
			Expect(dirID).NotTo(Equal(""))
		})
	})

	Describe("CreateProjectDirectories", func() {

	})

	Describe("CreateProject", func() {

	})
})
