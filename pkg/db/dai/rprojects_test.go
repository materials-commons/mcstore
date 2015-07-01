package dai

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/testdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RProjects", func() {
	var (
		rprojs Projects = NewRProjects(testdb.RSessionMust())
	)
	Describe("ForUser", func() {
		Context("ownedOnly", func() {
			It("Should return an error if owner doesn't have any projects", func() {
				projects, err := rprojs.ForUser("test1@mc.org", OwnedProjects)
				Expect(err).To(Equal(app.ErrNotFound))
				Expect(projects).To(BeNil())
			})

			It("Should return the project owned by user", func() {
				projects, err := rprojs.ForUser("test@mc.org", OwnedProjects)
				Expect(err).To(BeNil())
				Expect(projects).To(HaveLen(1))
			})
		})

		Context("all", func() {
			It("Should return an error if owner doesn't have any projects", func() {
				projects, err := rprojs.ForUser("does-not-exist", AllProjects)
				Expect(err).To(Equal(app.ErrNotFound))
				Expect(projects).To(BeNil())
			})

			It("Should return a project user has access to but doesn't own", func() {
				projects, err := rprojs.ForUser("test1@mc.org", AllProjects)
				Expect(err).To(BeNil())
				Expect(projects).To(HaveLen(1))
			})

			It("Should return both owned and accessible projects", func() {
				projects, err := rprojs.ForUser("test@mc.org", AllProjects)
				Expect(err).To(BeNil())
				Expect(projects).To(HaveLen(2))
			})
		})
	})
})
