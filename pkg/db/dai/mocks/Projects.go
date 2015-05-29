package mocks

import (
	"fmt"

	"github.com/materials-commons/testify/mock"
)

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

func (m *Projects) ByName(name, owner string) (*schema.Project, error) {
	ret := m.Called(name, owner)

	r0 := ret.Get(0).(*schema.Project)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *Projects) Insert(project *schema.Project) (*schema.Project, error) {
	ret := m.Called(project)
	r0 := ret.Get(0).(*schema.Project)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Projects) HasDirectory(projectID, directoryID string) bool {
	ret := m.Called(projectID, directoryID)
	r0 := ret.Get(0).(bool)
	return r0
}

func (m *Projects) AccessList(projectID string) ([]schema.Access, error) {
	ret := m.Called(projectID)
	r0 := ret.Get(0).([]schema.Access)
	r1 := ret.Error(1)
	return r0, r1
}

type pentry struct {
	project *schema.Project
	hasDir  bool
	err     error
	access  []schema.Access
}

type Projects2 struct {
	method        map[string]*pentry
	currentMethod string
}

func NewMProjects2() *Projects2 {
	return &Projects2{
		method: make(map[string]*pentry),
	}
}

func (m *Projects2) lookup(method string) *pentry {
	if e, ok := m.method[method]; ok {
		return e
	}
	panic(fmt.Sprintf("Unable to find method: %s", method))
}

func (m *Projects2) ByID(id string) (*schema.Project, error) {
	e := m.lookup("ByID")
	return e.project, e.err
}

func (m *Projects2) ByName(name, owner string) (*schema.Project, error) {
	e := m.lookup("ByName")
	return e.project, e.err
}

func (m *Projects2) Insert(project *schema.Project) (*schema.Project, error) {
	e := m.lookup("Insert")
	return e.project, e.err
}

func (m *Projects2) HasDirectory(projectID, directoryID string) bool {
	e := m.lookup("HasDirectory")
	return e.hasDir
}

func (m *Projects2) AccessList(projectID string) ([]schema.Access, error) {
	e := m.lookup("AccessList")
	return e.access, e.err
}

func (m *Projects2) On(method string) *Projects2 {
	m.currentMethod = method
	m.method[method] = &pentry{}
	return m
}

func (m *Projects2) SetError(err error) *Projects2 {
	m.method[m.currentMethod].err = err
	return m
}

func (m *Projects2) SetProject(project *schema.Project) *Projects2 {
	m.method[m.currentMethod].project = project
	return m
}

func (m *Projects2) SetHasDir(hasDir bool) *Projects2 {
	m.method[m.currentMethod].hasDir = hasDir
	return m
}

func (m *Projects2) SetAccessList(access []schema.Access) *Projects2 {
	m.method[m.currentMethod].access = access
	return m
}
