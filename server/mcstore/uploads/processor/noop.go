package processor

// noopFileProcessor does nothing. It exists so that the
// process method can be called on any file regardless of
// its type.
type noopFileProcessor struct {
}

// process does nothing. It exists so that process can be called on
// any type of file.
func (n *noopFileProcessor) Process() error {
	return nil
}
