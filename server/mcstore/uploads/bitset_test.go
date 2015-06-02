package uploads

import (
	"bytes"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/willf/bitset"
)

var _ = Describe("BitSets", func() {
	Describe("BitSet Serialization/Deserialization", func() {
		Context("JSON", func() {
			It("Should marshal and unmarshal JSON", func() {
				bset := bitset.New(5120)
				Expect(bset).NotTo(BeNil())
				Expect(bset.Count()).To(BeNumerically("==", 0))
				bset.Set(5)
				Expect(bset.Count()).To(BeNumerically("==", 1))
				b, err := bset.MarshalJSON()
				Expect(err).To(BeNil())
				Expect(b).NotTo(BeNil())
				var bset2 bitset.BitSet
				err = bset2.UnmarshalJSON(b)
				Expect(err).To(BeNil())
				Expect(bset2.Count()).To(BeNumerically("==", 1))
			})
		})

		Context("Stream", func() {
			It("Should work writing/reading from a byte buffer", func() {
				var b bytes.Buffer
				bset := bitset.New(5120)
				bset.Set(5)
				_, err := bset.WriteTo(&b)
				Expect(err).To(BeNil())
				var bset2 bitset.BitSet
				_, err = bset2.ReadFrom(&b)
				Expect(err).To(BeNil())
				Expect(bset2.Test(5)).To(BeTrue())
				Expect(bset2.Count()).To(BeNumerically("==", 1))
			})

			It("Should work writing/reading from a file", func() {
				f, err := os.Create("/tmp/bitset.out")
				Expect(err).To(BeNil())
				bset := bitset.New(5120)
				bset.Set(5)
				bset.Set(5119)
				_, err = bset.WriteTo(f)
				f.Close()
				f, err = os.Open("/tmp/bitset.out")
				Expect(err).To(BeNil())
				var bset2 bitset.BitSet
				_, err = bset2.ReadFrom(f)
				Expect(err).To(BeNil())
				Expect(bset2.Test(5)).To(BeTrue())
				Expect(bset2.Count()).To(BeNumerically("==", 2))
				Expect(bset2.Test(5119)).To(BeTrue())
			})
		})

		Context("misc", func() {
			It("Should round properly", func() {
				var _ = float64(2) / float64(2)
				//fmt.Println("x = ", int(math.Ceil(x)))
			})
		})
	})
})
