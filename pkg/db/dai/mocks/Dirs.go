package mocks

import "github.com/materials-commons/testify/mock"

import "github.com/materials-commons/mcstore/pkg/db/schema"

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

type Dirs2 struct {
	dir   *schema.Directory
	err   error
	files []schema.File
}

func NewMDirs2() *Dirs2 {
	return &Dirs2{}
}

func (m *Dirs2) ByID(id string) (*schema.Directory, error) {
	return m.dir, m.err
}

func (m *Dirs2) ByPath(path, projectID string) (*schema.Directory, error) {
	return m.dir, m.err
}

func (m *Dirs2) Files(dirID string) ([]schema.File, error) {
	return m.files, m.err
}

func (m *Dirs2) Insert(dir *schema.Directory) (*schema.Directory, error) {
	return m.dir, m.err
}

func (m *Dirs2) SetError(err error) {
	m.err = err
}

func (m *Dirs2) SetDir(dir *schema.Directory) {
	m.dir = dir
}

func (m *Dirs2) SetFiles(files []schema.File) {
	m.files = files
}
