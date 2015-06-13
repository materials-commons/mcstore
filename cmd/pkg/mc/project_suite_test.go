package mc

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MC Suite")
}
