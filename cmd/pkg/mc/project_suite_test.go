package mc

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"testing"
)

func TestMC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MC Suite")
}

var _ = BeforeSuite(func() {
	os.RemoveAll(".materialscommons")
	os.RemoveAll("/tmp/mcdir")
})

var _ = AfterSuite(func() {
	//	os.RemoveAll(".materialscommons")
	//	os.RemoveAll("/tmp/mcdir")
})
