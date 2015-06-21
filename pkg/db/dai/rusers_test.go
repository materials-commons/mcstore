package dai

import (
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/testdb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RUsers", func() {
	var rusers Users

	BeforeEach(func() {
		rusers = NewRUsers(testdb.RSession())
	})

	Describe("ByID", func() {
		It("Should find existing user", func() {
			u, err := rusers.ByID("test@mc.org")
			Expect(err).To(BeNil())
			Expect(u.ID).To(Equal("test@mc.org"))
		})

		It("Should return ErrNotFound when user doesn't exist", func() {
			u, err := rusers.ByID("does@not.exist")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(u).To(BeNil())
		})
	})

	Describe("ByAPIKey", func() {
		It("Should find user by APIKey", func() {
			u, err := rusers.ByAPIKey("test")
			Expect(err).To(BeNil())
			Expect(u.ID).To(Equal("test@mc.org"))
		})

		It("Should return ErrNotFound when apikey cannot be found", func() {
			u, err := rusers.ByAPIKey("no-such-key")
			Expect(err).To(Equal(app.ErrNotFound))
			Expect(u).To(BeNil())
		})
	})
})
