package mocks

import "github.com/materials-commons/testify/mock"

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Uploads struct {
	mock.Mock
}

func NewMUploads() *Uploads {
	return &Uploads{}
}

func (m *Uploads) ByID(id string) (*schema.Upload, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.Upload)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Uploads) Insert(upload *schema.Upload) (*schema.Upload, error) {
	ret := m.Called(upload)

	r0 := ret.Get(0).(*schema.Upload)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Uploads) Update(upload *schema.Upload) error {
	ret := m.Called(upload)

	r0 := ret.Error(0)

	return r0
}

func (m *Uploads) ForUser(user string) ([]schema.Upload, error) {
	ret := m.Called(user)

	r0 := ret.Get(0).([]schema.Upload)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Uploads) Delete(uploadID string) error {
	ret := m.Called(uploadID)

	r0 := ret.Error(0)

	return r0
}
