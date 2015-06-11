package mc

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mc Suite")
}
