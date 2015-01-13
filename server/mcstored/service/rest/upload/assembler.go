package upload

import "io"

// A Item contains data to assemble.
type Item interface {
	Name() string               // Name of the item
	Reader() (io.Reader, error) // Returns a reader to get at the items data
}

// A Finisher implements the method to call when assembly has finished successfully.
type Finisher interface {
	Finish() error
}

// A Assembler takes a list of items and assembles them.
type Assembler struct {
	items []Item
	Finisher
}

type ItemLess func(item1, item2 Item) bool

func NewAssembler(items []Item, finisher Finisher) *Assembler {
	return &Assembler{
		items:    items,
		Finisher: finisher,
	}
}

func (a *Assembler) To(destination io.Writer) error {
	err := writeEach(a.items, destination)
	switch {
	case err != nil:
		return err
	default:
		return a.Finish()
	}
}

func writeEach(items []Item, destination io.Writer) error {
	for _, item := range items {
		if err := writeItemTo(item, destination); err != nil {
			return err
		}
	}
	return nil
}

func writeItemTo(item Item, destination io.Writer) error {
	source, err := item.Reader()
	switch {
	case err != nil:
		return err
	default:
		_, err = io.Copy(destination, source)
		sclose, ok := source.(io.ReadCloser)
		if ok {
			sclose.Close()
		}
		return err
	}
}
