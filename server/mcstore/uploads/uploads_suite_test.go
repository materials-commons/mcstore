package uploads

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUploads(t *testing.T) {
	fmt.Println("TestUploads")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Uploads Suite")
}
