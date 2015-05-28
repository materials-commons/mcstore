package app

import (
	"github.com/materials-commons/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MCDir", func() {
	var saved string
	BeforeEach(func() {
		saved = MCDir.Path()
		config.Set("MCDIR", "/tmp/mcdir")
	})

	AfterEach(func() {
		config.Set("MCDIR", saved)
	})
	Describe("Path Method Tests", func() {
		It("Should panic when MCDIR is not set", func() {
			config.Set("MCDIR", "")
			f := func() {
				MCDir.Path()
			}
			Expect(f).To(Panic())
		})

		It("Should return the value of setting MCDIR", func() {
			Expect(MCDir.Path()).To(Equal("/tmp/mcdir"))
		})
	})

	Describe("FileDir Method Tests", func() {
		It("Should return path for good file id", func() {
			fileID := "abc-defg-ghi-jkl-mnopqr"
			dir := MCDir.FileDir(fileID)
			Expect(dir).To(Equal("/tmp/mcdir/de/fg"))
		})

		It("Should return an empty string for a bad id", func() {
			fileID := "bad_file_id"
			dir := MCDir.FileDir(fileID)
			Expect(dir).To(Equal(""))
		})

		It("Should return an empty string if there is a bad segment in the id", func() {
			fileID := "abc-def-ghi-jkl"
			dir := MCDir.FileDir(fileID)
			Expect(dir).To(Equal(""))
		})
	})

	Describe("FilePath Method Tests", func() {
		It("Should return path for a good file id", func() {
			fileID := "abc-defg-ghi-jkl-mnopqr"
			path := MCDir.FilePath(fileID)
			Expect(path).To(Equal("/tmp/mcdir/de/fg/abc-defg-ghi-jkl-mnopqr"))
		})

		It("Should return an empty string for a bad id", func() {
			fileID := "bad_file_id"
			path := MCDir.FilePath(fileID)
			Expect(path).To(Equal(""))
		})

		It("Should return an empty string if there is a bad segment in the id", func() {
			fileID := "abc-def-ghi-jkl"
			path := MCDir.FilePath(fileID)
			Expect(path).To(Equal(""))
		})
	})

	Describe("FileConversionDir Method Tests", func() {
		It("Should return path for a good file id", func() {
			fileID := "abc-defg-ghi"
			path := MCDir.FileConversionDir(fileID)
			Expect(path).To(Equal("/tmp/mcdir/de/fg/.conversion"))
		})

		It("Should return an empty string for a bad id", func() {
			fileID := "bad_file_id"
			dir := MCDir.FileConversionDir(fileID)
			Expect(dir).To(Equal(""))
		})

		It("Should return an empty string if there is a bad segment in the id", func() {
			fileID := "abc-def-ghi-jkl"
			dir := MCDir.FileConversionDir(fileID)
			Expect(dir).To(Equal(""))
		})
	})

	Describe("FilePathImageConversion Method Tests", func() {
		It("Should return path for a good file id", func() {
			fileID := "abc-defg-ghi"
			path := MCDir.FilePathImageConversion(fileID)
			Expect(path).To(Equal("/tmp/mcdir/de/fg/.conversion/abc-defg-ghi.jpg"))
		})

		It("Should return an empty string for a bad id", func() {
			fileID := "bad_file_id"
			path := MCDir.FilePathImageConversion(fileID)
			Expect(path).To(Equal(""))
		})

		It("Should return an empty string if there is a bad segment in the id", func() {
			fileID := "abc-def-ghi-jkl"
			path := MCDir.FilePathImageConversion(fileID)
			Expect(path).To(Equal(""))
		})
	})

	Describe("UploadDir Method Tests", func() {
		It("Should return path for a good upload id", func() {
			uploadID := "abc-defg-ghi"
			path := MCDir.UploadDir(uploadID)
			Expect(path).To(Equal("/tmp/mcdir/upload/abc-defg-ghi"))
		})

		It("Should return a path even for malformed ids", func() {
			uploadID := "bad_upload_id"
			dir := MCDir.UploadDir(uploadID)
			Expect(dir).To(Equal("/tmp/mcdir/upload/bad_upload_id"))
		})
	})
})
