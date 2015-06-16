package uploads

import "github.com/willf/bitset"

type tracker interface {
	setBlock(id string, block int)
	done(id string) bool
	load(id string, numBlocks int)
	clear(id string)
	clearBlock(id string, block int)
	hash(id string) string
	addToHash(id string, what []byte)
	getBlocks(id string) *bitset.BitSet
	isBlockSet(id string, block int) bool
	idExists(id string) bool
}

var (
	requestBlockTracker tracker = newBlockTracker()
)
