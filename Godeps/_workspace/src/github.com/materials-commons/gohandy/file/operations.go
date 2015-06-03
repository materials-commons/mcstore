package file

import "os"

// Operations is the operations on a file. The interface
// exists to allow mocking of file system operations.
type Operations interface {
	Remove(path string) error
	RemoveAll(path string) error
	Mkdir(path string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Create(path string) (*os.File, error)
	Open(name string) (file *os.File, err error)
}

// osOperations implements the os package operations
type osOperations struct{}

// OS allows access to the os package methods that are the
// same as in the Operations interface.
var OS osOperations

// Remove is a wrapper around os.Remove
func (_ osOperations) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll is a wrapper around os.RemoveAll
func (_ osOperations) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Mkdir is a wrapper around os.Mkdir
func (_ osOperations) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(path, perm)
}

// MkdirAll is a wrapper around os.MkdirAll
func (_ osOperations) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Create is a wrapper around os.Create
func (_ osOperations) Create(path string) (*os.File, error) {
	return os.Create(path)
}

// Open is a wrapper around os.Open
func (_ osOperations) Open(path string) (*os.File, error) {
	return os.Open(path)
}

type MockOperations struct {
	Err error
}

var MockOps MockOperations = MockOperations{}

// Remove returns Err from MockOps. It does nothing.
func (m MockOperations) Remove(path string) error {
	return m.Err
}

// RemoveAll returns Err from MockOps. It does nothing.
func (m MockOperations) RemoveAll(path string) error {
	return m.Err
}

// Mkdir returns Err from MockOps. It does nothing.
func (m MockOperations) Mkdir(path string, perm os.FileMode) error {
	return m.Err
}

// MkdirAll returns Err from MockOps. It does nothing.
func (m MockOperations) MkdirAll(path string, perm os.FileMode) error {
	return m.Err
}

// Create returns nil, Err from MockOps if MockOps Err is
// not nil. Otherwise it returns os.Create(os.DevNull).
func (m MockOperations) Create(path string) (*os.File, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return os.Create(os.DevNull)
}

// Open returns nil, Err from MockOps if MockOps err is not
// nil. Otherwise it returns os.Open(os.DevNull).
func (m MockOperations) Open(path string) (*os.File, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return os.Open(os.DevNull)
}