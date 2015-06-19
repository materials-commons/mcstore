package mocks

import "github.com/materials-commons/testify/mock"

import (
	"fmt"
	"github.com/materials-commons/mcstore/pkg/db/schema"
)

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

func (m *Files) AllByChecksum(checksum string) ([]schema.File, error) {
	ret := m.Called(checksum)
	r0 := ret.Get(0).([]schema.File)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Files) ByPath(name, dirID string) (*schema.File, error) {
	ret := m.Called(name, dirID)
	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Files) Insert(file *schema.File, dirID string, projectID string) (*schema.File, error) {
	ret := m.Called()
	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Files) Update(file *schema.File) error {
	ret := m.Called()
	r0 := ret.Error(0)
	return r0
}

func (m *Files) UpdateFields(fileID string, fields map[string]interface{}) error {
	ret := m.Called(fileID)
	r0 := ret.Error(0)
	return r0
}

func (m *Files) Delete(fileID, directoryID, projectID string) (*schema.File, error) {
	ret := m.Called(fileID, directoryID, projectID)
	r0 := ret.Get(0).(*schema.File)
	r1 := ret.Error(1)
	return r0, r1
}

func (m *Files) GetProject(fileID string) (*schema.Project, error) {
	ret := m.Called(fileID)
	r0 := ret.Get(0).(*schema.Project)
	r1 := ret.Error(1)
	return r0, r1
}

type fentry struct {
	file    *schema.File
	err     error
	project *schema.Project
	files   []schema.File
}

type Files2 struct {
	method        map[string]*fentry
	currentMethod string
}

func NewMFiles2() *Files2 {
	return &Files2{
		method: make(map[string]*fentry),
	}
}

func (m *Files2) lookup(method string) *fentry {
	if e, ok := m.method[method]; ok {
		return e
	}
	panic(fmt.Sprintf("Unable to find method: %s", method))
}

func (m *Files2) ByID(id string) (*schema.File, error) {
	e := m.lookup("ByID")
	return e.file, e.err
}

func (m *Files2) ByChecksum(checksum string) (*schema.File, error) {
	e := m.lookup("ByChecksum")
	return e.file, e.err
}

func (m *Files2) AllByChecksum(checksum string) ([]schema.File, error) {
	e := m.lookup("AllByChecksum")
	return e.files, e.err
}

func (m *Files2) ByPath(name, dirID string) (*schema.File, error) {
	e := m.lookup("ByPath")
	return e.file, e.err
}

func (m *Files2) Insert(file *schema.File, dirID string, projectID string) (*schema.File, error) {
	e := m.lookup("Insert")
	return e.file, e.err
}

func (m *Files2) Update(file *schema.File) error {
	e := m.lookup("Update")
	return e.err
}

func (m *Files2) UpdateFields(fileID string, fields map[string]interface{}) error {
	e := m.lookup("UpdateFields")
	return e.err
}

func (m *Files2) Delete(fileID, directoryID, projectID string) (*schema.File, error) {
	e := m.lookup("Delete")
	return e.file, e.err
}

func (m *Files2) GetProject(fileID string) (*schema.Project, error) {
	e := m.lookup("GetProject")
	return e.project, e.err
}

func (m *Files2) On(method string) *Files2 {
	m.currentMethod = method
	m.method[method] = &fentry{}
	return m
}

func (m *Files2) SetError(err error) *Files2 {
	m.method[m.currentMethod].err = err
	return m
}

func (m *Files2) SetFile(file *schema.File) *Files2 {
	m.method[m.currentMethod].file = file
	return m
}

func (m *Files2) SetFiles(files []schema.File) *Files2 {
	m.method[m.currentMethod].files = files
	return m
}

func (m *Files2) SetProject(project *schema.Project) *Files2 {
	m.method[m.currentMethod].project = project
	return m
}
