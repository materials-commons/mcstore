package file

import (
	"os"

	"time"
)

// Operations is the operations on a file. The interface
// exists to allow mocking of file system operations.
type Operations interface {
	Remove(path string) error
	RemoveAll(path string) error
	Mkdir(path string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Create(path string) (*os.File, error)
	Open(name string) (file *os.File, err error)
	Rename(oldpath, newpath string) error
	Stat(path string) (os.FileInfo, error)
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

// Rename is a wrapper around os.Rename
func (_ osOperations) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Stat is a wrapper around os.Stat
func (_ osOperations) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

type MockFileOpEntry struct {
	Err    error
	RValue interface{}
}

type MockOperations struct {
	method        map[string]*MockFileOpEntry
	currentMethod string
}

func MockOps() *MockOperations {
	return &MockOperations{
		method: make(map[string]*MockFileOpEntry),
	}
}

func (m *MockOperations) lookup(method string) *MockFileOpEntry {
	if e, ok := m.method[method]; ok {
		return e
	}
	return nil
}

// Remove returns Err from MockOps. It does nothing.
func (m *MockOperations) Remove(path string) error {
	if entry := m.lookup("Remove"); entry != nil {
		return entry.Err
	}
	return nil
}

// RemoveAll returns Err from MockOps. It does nothing.
func (m *MockOperations) RemoveAll(path string) error {
	if entry := m.lookup("RemoveAll"); entry != nil {
		return entry.Err
	}
	return nil
}

// Mkdir returns Err from MockOps. It does nothing.
func (m *MockOperations) Mkdir(path string, perm os.FileMode) error {
	if entry := m.lookup("Mkdir"); entry != nil {
		return entry.Err
	}
	return nil
}

// MkdirAll returns Err from MockOps. It does nothing.
func (m *MockOperations) MkdirAll(path string, perm os.FileMode) error {
	if entry := m.lookup("MkdirAll"); entry != nil {
		return entry.Err
	}
	return nil
}

// Create returns nil, Err from MockOps if MockOps Err is
// not nil. Otherwise it returns os.Create(os.DevNull).
func (m *MockOperations) Create(path string) (*os.File, error) {
	if entry := m.lookup("Create"); entry != nil {
		return nil, entry.Err
	}
	return os.Create(os.DevNull)
}

// Open returns nil, Err from MockOps if MockOps err is not
// nil. Otherwise it returns os.Open(os.DevNull).
func (m *MockOperations) Open(path string) (*os.File, error) {
	if entry := m.lookup("Open"); entry != nil {
		return nil, entry.Err
	}
	return os.Open(os.DevNull)
}

// Rename returns Err from MockOps. It does nothing.
func (m *MockOperations) Rename(oldpath, newpath string) error {
	if entry := m.lookup("Rename"); entry != nil {
		return entry.Err
	}
	return nil
}

type MockFileInfo struct {
	MName    string
	MSize    int64
	MMode    os.FileMode
	MModTime time.Time
	MIsDir   bool
	MSys     interface{}
}

func (fi MockFileInfo) Name() string {
	return fi.MName
}

func (fi MockFileInfo) Size() int64 {
	return fi.MSize
}

func (fi MockFileInfo) Mode() os.FileMode {
	return fi.MMode
}

func (fi MockFileInfo) ModTime() time.Time {
	return fi.MModTime
}

func (fi MockFileInfo) IsDir() bool {
	return fi.MIsDir
}

func (fi MockFileInfo) Sys() interface{} {
	return fi.MSys
}

func (m *MockOperations) Stat(path string) (os.FileInfo, error) {
	if entry := m.lookup("Stat"); entry != nil {
		rv := entry.RValue.(MockFileInfo)
		return rv, entry.Err
	}

	return nil, os.ErrNotExist
}

func (m *MockOperations) On(method string) *MockOperations {
	m.currentMethod = method
	m.method[method] = &MockFileOpEntry{}
	return m
}

func (m *MockOperations) SetError(err error) *MockOperations {
	m.method[m.currentMethod].Err = err
	return m
}

func (m *MockOperations) SetValue(what interface{}) *MockOperations {
	m.method[m.currentMethod].RValue = what
	return m
}
