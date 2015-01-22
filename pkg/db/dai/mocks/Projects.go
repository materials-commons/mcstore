package mocks

import "github.com/materials-commons/testify/mock"

import "github.com/materials-commons/mcstore/pkg/db/schema"

type Projects struct {
	mock.Mock
}

func NewMProjects() *Projects {
	return &Projects{}
}

func (m *Projects) ByID(id string) (*schema.Project, error) {
	ret := m.Called(id)

	r0 := ret.Get(0).(*schema.Project)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Projects) HasDirectory(projectID, directoryID string) bool {
	ret := m.Called(projectID, directoryID)
	r0 := ret.Get(0).(bool)
	return r0
}