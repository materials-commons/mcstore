package uploads

import (
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/app/flow"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestPath", func() {
	var (
		rpath          mcdirRequestPath
		savedMCDIRPath string
		req            flow.Request
	)

	BeforeEach(func() {
		savedMCDIRPath = app.MCDir.Path()
		config.Set("MCDIR", "/tmp")
		req = flow.Request{
			FlowIdentifier:  "testid",
			FlowChunkNumber: 1,
		}
	})

	AfterEach(func() {
		config.Set("MCDIR", savedMCDIRPath)
	})

	Context("Path", func() {
		It("Should return a valid path", func() {
			requestPath := rpath.path(&req)
			Expect(requestPath).To(Equal("/tmp/upload/testid/1"))
		})
	})

	Context("Dir", func() {
		It("Should return a valid dir", func() {
			requestDir := rpath.dir(&req)
			Expect(requestDir).To(Equal("/tmp/upload/testid"))
		})
	})
})
