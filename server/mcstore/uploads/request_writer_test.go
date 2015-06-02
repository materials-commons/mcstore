package uploads

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/materials-commons/mcstore/pkg/app/flow"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("requestWriter", func() {
	Describe("blockRequestWriter", func() {
		var (
			brWriter *blockRequestWriter = &blockRequestWriter{}
			dirPath  string              = filepath.Join(os.TempDir(), "rr")
			filePath string              = filepath.Join(dirPath, "req")
			req      *flow.Request       = &flow.Request{
				FlowIdentifier:  "req",
				FlowChunkNumber: 1,
			}
		)

		BeforeEach(func() {
			os.MkdirAll(dirPath, 0770)
		})

		AfterEach(func() {
			//os.RemoveAll(dirPath)
		})

		Context("write method tests", func() {
			It("Should correctly write a single block file", func() {
				data := []byte("hello")
				req.Chunk = data
				req.FlowTotalSize = int64(len(data))
				err := brWriter.write(dirPath, req)
				Expect(err).To(BeNil(), "Error = %s", err)
				content, err := ioutil.ReadFile(filePath)
				Expect(err).To(BeNil())
				Expect(content).To(Equal(data))
			})

		})

		Context("write method tests", func() {
			It("Should fail", func() {
				Expect("").To(Equal(""))
			})
		})

		Context("validate method tests", func() {

		})
	})
})
