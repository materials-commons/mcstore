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

// NewAssembler creates an Assembler.
func NewAssembler(items []Item, finisher Finisher) *Assembler {
	return &Assembler{
		items:    items,
		Finisher: finisher,
	}
}

// To will write the assembled items to destination. It writes
// the items in the order they were give. If it can't write
// any item, it will quit on that item and return it's error.
// If it is able to write all items then it will call Finisher.
// It only calls Finisher if it was able to successfully write
// all items. If it calls Finisher it will return its result.
func (a *Assembler) To(destination io.Writer) error {
	err := writeEach(a.items, destination)
	switch {
	case err != nil:
		return err
	default:
		return a.Finish()
	}
}

// writeEach attempts to write each item to destination. It
// will stop on the first item it cannot write and return
// its error.
func writeEach(items []Item, destination io.Writer) error {
	for _, item := range items {
		if err := writeItemTo(item, destination); err != nil {
			return err
		}
	}
	return nil
}

// writeItemTo performs the write to destination of a particular
// item. It calls copy to append the item to destination. If the
// reader returned by a item is a ReadCloser then it will call
// the close routine.
func writeItemTo(item Item, destination io.Writer) error {
	source, err := item.Reader()
	switch {
	case err != nil:
		return err
	default:
		sclose, ok := source.(io.ReadCloser)
		if ok {
			defer sclose.Close()
		}
		_, err = io.Copy(destination, source)
		return err
	}
}
