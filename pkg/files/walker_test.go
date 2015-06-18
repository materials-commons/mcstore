package files

import (
	"os"
	"path/filepath"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Walker", func() {
	var (
		testPath string
		walker   PWalker
	)
	BeforeEach(func() {
		testPath = filepath.Join(os.TempDir(), "filetree-test")
		os.MkdirAll(testPath, 0777)
		f, _ := os.Create(filepath.Join(testPath, "f1.txt"))
		f.WriteString("empty")
		f.Close()

		os.MkdirAll(filepath.Join(testPath, "d2"), 0777)
		f, _ = os.Create(filepath.Join(testPath, "d2", "f2.txt"))
		f.WriteString("hello f2.txt")
		f.Close()

		walker = PWalker{
			NumParallel: 1,
			ProcessDirs: true,
		}
	})

	AfterEach(func() {
		os.RemoveAll(testPath)
	})

	It("Should not return an error", func() {
		fn := func(done <-chan struct{}, entries <-chan TreeEntry, result chan<- string) {
			for entry := range entries {
				select {
				case <-done:
					return
				default:
					fmt.Println("Path = ", entry.Path)
					fmt.Println("Name = ", entry.Finfo.Name())
					result <- ""
				}
			}
		}
		walker.ProcessFn = fn
		_, errc := walker.PWalk(testPath)
		err := <-errc
		Expect(err).To(BeNil())
	})
})
