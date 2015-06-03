package uploads

type tracker interface {
	setBlock(id string, block int)
	done(id string) bool
	load(id string, numBlocks int)
	clear(id string)
	clearBlock(id string, block int)
	hash(id string) string
	addToHash(id string, what []byte)
}

var (
	requestBlockCountTracker tracker = newBlockCountTracker()
	requestBlockTracker      tracker = newBlockTracker()
)
