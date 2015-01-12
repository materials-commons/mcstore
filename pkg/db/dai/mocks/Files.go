package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Files struct {
	mock.Mock
}

func NewMFiles() *Files {
	return &Files{}
}

func (m *Files) ByID(id string) (*schema.File, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
