// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// ExitHandler is an autogenerated mock type for the ExitHandler type
type ExitHandler struct {
	mock.Mock
}

// Error provides a mock function with given fields: _a0
func (_m *ExitHandler) Error(_a0 error) {
	_m.Called(_a0)
}

// Fail provides a mock function with given fields:
func (_m *ExitHandler) Fail() {
	_m.Called()
}

// Pass provides a mock function with given fields:
func (_m *ExitHandler) Pass() {
	_m.Called()
}

// NewExitHandler creates a new instance of ExitHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExitHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *ExitHandler {
	mock := &ExitHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
