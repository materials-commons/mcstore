package mocks

import "github.com/materials-commons/testify/mock"

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Access struct {
	mock.Mock
}

func NewMAccess() *Access {
	return &Access{}
}

func (m *Access) AllowedByOwner(owner, user string) bool {
	ret := m.Called(owner, user)

	r0 := ret.Get(0).(bool)

	return r0
}

func (m *Access) GetFile(apikey, fileID string) (*schema.File, error) {
	ret := m.Called(apikey, fileID)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}
