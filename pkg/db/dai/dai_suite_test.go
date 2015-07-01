package dai

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDai(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dai Suite")
}
