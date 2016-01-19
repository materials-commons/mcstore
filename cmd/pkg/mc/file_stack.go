package mc

type fileEntry struct {
	Type     string
	ID       string
	Path     string
	Size     int64
	Checksum string
}

type element struct {
	file *fileEntry
	next *element
}

type fileStack struct {
	top  *element
	size int
}

func (s *fileStack) Len() int {
	return s.size
}

func (s *fileStack) Push(f *fileEntry) {
	s.top = &element{f, s.top}
	s.size++
}

func (s *fileStack) Pop() (file *fileEntry) {
	if s.size > 0 {
		file, s.top = s.top.file, s.top.next
		s.size--
		return file
	}
	return nil
}

func (s *fileStack) Peek() (file *fileEntry, exists bool) {
	exists = false
	if s.size > 0 {
		file = s.top.file
		exists = true
	}
	return file, exists
}
