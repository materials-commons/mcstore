package uploads

import (
	"crypto/md5"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlockTracker", func() {
	var btracker *blockTracker

	BeforeEach(func() {
		btracker = newBlockTracker()
	})

	Describe("Hash Tests", func() {

		It("Should match for a single block hash", func() {
			btracker.load("abc", 1)
			btracker.addToHash("abc", []byte("hello"))
			expected := fmt.Sprintf("%x", md5.Sum([]byte("hello")))
			got := btracker.hash("abc")
			Expect(expected).To(Equal(got))
		})

		It("Should match for a multiple block hash", func() {
			btracker.load("abc", 2)
			btracker.addToHash("abc", []byte("hello"))
			btracker.addToHash("abc", []byte("world"))
			expected := fmt.Sprintf("%x", md5.Sum([]byte("helloworld")))
			got := btracker.hash("abc")
			Expect(expected).To(Equal(got))
		})
	})

	Describe("done method tests", func() {
		It("Should mark as done for single block tracker", func() {
			btracker.load("abc", 1)
			Expect(btracker.done("abc")).To(BeFalse())
			btracker.setBlock("abc", 1)
			Expect(btracker.done("abc")).To(BeTrue())
		})

		It("Should not be done if we explicitly clear a block", func() {
			btracker.load("abc", 1)
			btracker.setBlock("abc", 1)
			Expect(btracker.done("abc")).To(BeTrue())
			btracker.clearBlock("abc", 1)
			Expect(btracker.done("abc")).To(BeFalse())
		})

		It("Should be done only after all blocks are marked", func() {
			btracker.load("abc", 2)
			Expect(btracker.done("abc")).To(BeFalse())
			btracker.setBlock("abc", 1)
			Expect(btracker.done("abc")).To(BeFalse())
			btracker.setBlock("abc", 2)
			Expect(btracker.done("abc")).To(BeTrue())
		})
	})
})
