package mocks

import "github.com/stretchr/testify/mock"

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Users struct {
	mock.Mock
}

func NewMUsers() *Users {
	return &Users{}
}

func (m *Users) ByAPIKey(apikey string) (*schema.User, error) {
	ret := m.Called(apikey)

	r0 := ret.Get(0).(*schema.User)
	r1 := ret.Error(1)

	return r0, r1
}
