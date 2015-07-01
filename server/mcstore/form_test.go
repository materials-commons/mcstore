package mcstore

import (
	"bufio"
	"bytes"
	"mime/multipart"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Form", func() {
	It("Should correctly parse a good form", func() {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		fw := multipart.NewWriter(w)

		fw.WriteField("flowChunkNumber", "0")
		fw.WriteField("flowTotalChunks", "1")
		fw.WriteField("flowChunkSize", "5")
		fw.WriteField("flowTotalSize", "5")
		fw.WriteField("flowIdentifier", "unique")
		fw.WriteField("flowFileName", "test.txt")
		fw.WriteField("flowRelativePath", "test.txt")
		fw.WriteField("projectID", "project")
		fw.WriteField("directoryID", "directory")
		fw.WriteField("fileID", "file")
		fw.WriteField("chunkData", "hello")

		err := fw.Close()
		w.Flush() // Need to flush the writer.

		Expect(err).To(BeNil(), "Got unexpected error: %s", err)

		reader := bufio.NewReader(&b)
		req, err := multipart2FlowRequest(multipart.NewReader(reader, fw.Boundary()))
		Expect(err).To(BeNil(), "Got unexpected error: %s", err)
		Expect(req).NotTo(BeNil())

		Expect(req.FlowChunkNumber).To(BeNumerically("==", 0))
		Expect(req.FlowTotalChunks).To(BeNumerically("==", 1))
		Expect(req.FlowChunkSize).To(BeNumerically("==", 5))
		Expect(req.FlowTotalSize).To(BeNumerically("==", 5))
		Expect(req.FlowIdentifier).To(Equal("unique"))
		Expect(req.FlowFileName).To(Equal("test.txt"))
		Expect(req.FlowRelativePath).To(Equal("test.txt"))
		Expect(req.ProjectID).To(Equal("project"))
		Expect(req.DirectoryID).To(Equal("directory"))
		Expect(req.FileID).To(Equal("file"))

		s := string(req.Chunk)
		Expect(s).To(Equal("hello"))
	})
})
