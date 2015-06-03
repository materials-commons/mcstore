package uploads

type tracker interface {
	setBlock(id string, block int)
	done(id string) bool
	setup(id string, numBlocks int)
	clear(id string)
	clearBlock(id string, block int)
}

var (
	requestBlockCountTracker tracker = newBlockCountTracker()
	requestBlockTracker      tracker = newBlockTracker()
)
