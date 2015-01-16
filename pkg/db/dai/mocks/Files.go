package mocks

import "github.com/materials-commons/testify/mock"

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

func (m *Files) ByChecksum(checksum string) (*schema.File, error) {
	ret := m.Called(checksum)

	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Files) Insert(file *schema.File, dirID string, projectID string) (*schema.File, error) {
	ret := m.Called(file, dirID, projectID)
	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Files) Update(file *schema.File) error {
	ret := m.Called(file)
	r0 := ret.Error(0)
	return r0
}

func (m *Files) UpdateFields(fileID string, fields map[string]interface{}) error {
	ret := m.Called(fileID, fields)
	r0 := ret.Error(0)
	return r0
}
