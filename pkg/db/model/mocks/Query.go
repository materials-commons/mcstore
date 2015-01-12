package mocks

import "github.com/materials-commons/testify/mock"

import r "github.com/dancannon/gorethink"

type Query struct {
	mock.Mock
}

func (m *Query) Rql() r.Term {
	ret := m.Called()

	r0 := ret.Get(0).(r.Term)

	return r0
}
func (m *Query) Session() *r.Session {
	ret := m.Called()

	r0 := ret.Get(0).(*r.Session)

	return r0
}
func (m *Query) ByID(id string, obj interface{}) error {
	ret := m.Called(id, obj)

	r0 := ret.Error(0)

	return r0
}
func (m *Query) Row(query r.Term, obj interface{}) error {
	ret := m.Called(query, obj)

	r0 := ret.Error(0)

	return r0
}
func (m *Query) Rows(query r.Term, results interface{}) error {
	ret := m.Called(query, results)

	r0 := ret.Error(0)

	return r0
}
func (m *Query) Update(id string, what interface{}) error {
	ret := m.Called(id, what)

	r0 := ret.Error(0)

	return r0
}
func (m *Query) InsertRaw(table string, what interface{}, dest interface{}) error {
	ret := m.Called(table, what, dest)

	r0 := ret.Error(0)

	return r0
}
func (m *Query) Insert(what interface{}, dest interface{}) error {
	ret := m.Called(what, dest)

	r0 := ret.Error(0)

	return r0
}
func (m *Query) Delete(id string) error {
	ret := m.Called(id)

	r0 := ret.Error(0)

	return r0
}
