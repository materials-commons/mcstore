package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Groups struct {
	mock.Mock
}

func NewMGroups() *Groups {
	return &Groups{}
}

func (m *Groups) ByID(id string) (*schema.Group, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.Group)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Groups) ForOwner(owner string) ([]schema.Group, error) {
	ret := m.Called(owner)

	r0 := ret.Get(0).([]schema.Group)
	r1 := ret.Error(1)

	return r0, r1
}
