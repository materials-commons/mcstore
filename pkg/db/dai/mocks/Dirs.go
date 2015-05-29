package mocks

import "github.com/materials-commons/testify/mock"

import (
	"fmt"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

type Dirs struct {
	mock.Mock
}

func NewMDirs() *Dirs {
	return &Dirs{}
}

func (m *Dirs) ByID(id string) (*schema.Directory, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.Directory)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Dirs) ByPath(path, projectID string) (*schema.Directory, error) {
	ret := m.Called(path, projectID)
	r0 := ret.Get(0).(*schema.Directory)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Dirs) Files(dirID string) ([]schema.File, error) {
	ret := m.Called(dirID)
	r0 := ret.Get(0).([]schema.File)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Dirs) Insert(dir *schema.Directory) (*schema.Directory, error) {
	ret := m.Called(dir)
	r0 := ret.Get(0).(*schema.Directory)
	r1 := ret.Error(1)
	return r0, r1
}

type entry struct {
	dir   *schema.Directory
	err   error
	files []schema.File
}

type Dirs2 struct {
	method        map[string]*entry
	currentMethod string
}

func NewMDirs2() *Dirs2 {
	return &Dirs2{
		method: make(map[string]*entry),
	}
}

func (m *Dirs2) lookup(method string) *entry {
	if e, ok := m.method[method]; ok {
		return e
	}
	panic(fmt.Sprintf("Unable to find method: %s", method))
}

func (m *Dirs2) ByID(id string) (*schema.Directory, error) {
	e := m.lookup("ByID")
	return e.dir, e.err
}

func (m *Dirs2) ByPath(path, projectID string) (*schema.Directory, error) {
	e := m.lookup("ByPath")
	return e.dir, e.err
}

func (m *Dirs2) Files(dirID string) ([]schema.File, error) {
	e := m.lookup("Files")
	return e.files, e.err
}

func (m *Dirs2) Insert(dir *schema.Directory) (*schema.Directory, error) {
	e := m.lookup("Insert")
	return e.dir, e.err
}

func (m *Dirs2) On(method string) *Dirs2 {
	m.currentMethod = method
	m.method[method] = &entry{}
	return m
}

func (m *Dirs2) SetError(err error) *Dirs2 {
	m.method[m.currentMethod].err = err
	return m
}

func (m *Dirs2) SetDir(dir *schema.Directory) *Dirs2 {
	m.method[m.currentMethod].dir = dir
	return m
}

func (m *Dirs2) SetFiles(files []schema.File) *Dirs2 {
	m.method[m.currentMethod].files = files
	return m
}
