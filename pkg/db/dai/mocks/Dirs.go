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
