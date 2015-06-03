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
			req      *flow.Request
		)

		BeforeEach(func() {
			os.MkdirAll(dirPath, 0770)
			req = &flow.Request{
				FlowIdentifier:  "req",
				FlowChunkNumber: 1,
			}
		})

		AfterEach(func() {
			os.RemoveAll(dirPath)
		})

		Context("write method tests", func() {
			It("Should correctly write a single block file", func() {
				data := []byte("hello")
				req.Chunk = data
				req.FlowChunkNumber = int32(len(data))
				req.FlowTotalSize = int64(len(data))
				err := brWriter.write(dirPath, req)
				Expect(err).To(BeNil(), "Error = %s", err)
				content, err := ioutil.ReadFile(filePath)
				Expect(err).To(BeNil())
				Expect(content).To(Equal(data))
			})

			It("Should correctly write a multiple block file", func() {
				data1 := []byte("my")
				data2 := []byte("da")
				req.Chunk = data1
				req.FlowChunkSize = 2
				req.FlowTotalSize = int64(len(data1) + len(data2))
				err := brWriter.write(dirPath, req)
				Expect(err).To(BeNil(), "error = %s", err)
				content, err := ioutil.ReadFile(filePath)
				Expect(err).To(BeNil())

				// The file has a length of 3, since we only wrote 2 and the other parts haven't
				// yet been written, we just compare the FlowChunkSize (2) bytes.
				Expect(content[:req.FlowChunkSize]).To(Equal(data1), "not equal '%s'/'%s'", string(content), string(data1))
				req.Chunk = data2
				req.FlowChunkNumber = 2
				err = brWriter.write(dirPath, req)
				Expect(err).To(BeNil(), "error = %s", err)
				content, err = ioutil.ReadFile(filePath)
				Expect(err).To(BeNil())
				var dataCombined []byte
				dataCombined = append(dataCombined, data1...)
				dataCombined = append(dataCombined, data2...)
				Expect(content).To(Equal(dataCombined))
			})
		})

		It("Should correctly write a multiple block file with the final block short", func() {
			data1 := []byte("my")
			data2 := []byte("d")
			req.Chunk = data1
			req.FlowChunkSize = 2
			req.FlowTotalSize = int64(len(data1) + len(data2))
			err := brWriter.write(dirPath, req)
			Expect(err).To(BeNil(), "error = %s", err)
			content, err := ioutil.ReadFile(filePath)
			Expect(err).To(BeNil())

			// The file has a length of 3, since we only wrote 2 and the other parts haven't
			// yet been written, we just compare the FlowChunkSize (2) bytes.
			Expect(content[:req.FlowChunkSize]).To(Equal(data1), "not equal '%s'/'%s'", string(content), string(data1))
			req.Chunk = data2
			req.FlowChunkNumber = 2
			err = brWriter.write(dirPath, req)
			Expect(err).To(BeNil(), "error = %s", err)
			content, err = ioutil.ReadFile(filePath)
			Expect(err).To(BeNil())
			var dataCombined []byte
			dataCombined = append(dataCombined, data1...)
			dataCombined = append(dataCombined, data2...)
			Expect(content).To(Equal(dataCombined))
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
