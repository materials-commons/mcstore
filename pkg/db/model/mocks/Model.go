package mocks

import "github.com/materials-commons/mcstore/pkg/db/model"
import "github.com/stretchr/testify/mock"

import r "github.com/dancannon/gorethink"

type Model struct {
	mock.Mock
}

func (m *Model) Q() model.Query {
	ret := m.Called()

	r0 := ret.Get(0).(model.Query)

	return r0
}
func (m *Model) Qs(session *r.Session) model.Query {
	ret := m.Called(session)

	r0 := ret.Get(0).(model.Query)

	return r0
}
func (m *Model) Table() r.Term {
	ret := m.Called()

	r0 := ret.Get(0).(r.Term)

	return r0
}
func (m *Model) T() r.Term {
	ret := m.Called()

	r0 := ret.Get(0).(r.Term)

	return r0
}
