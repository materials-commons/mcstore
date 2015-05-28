package mcstore

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMcstore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mcstore Suite")
}
