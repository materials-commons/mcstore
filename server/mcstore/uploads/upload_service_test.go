package uploads

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UploadService", func() {
	var (
		s UploadService
	)

	BeforeEach(func() {
		s = NewUploadService()
	})

	Describe("something", func() {
		Context("context", func() {
			It("Should do something", func() {
				str := ""
				Expect(str).To(Equal(""))
			})
		})
	})

})
