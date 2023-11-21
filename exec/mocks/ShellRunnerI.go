// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	bytes "bytes"
	exec "os/exec"

	mock "github.com/stretchr/testify/mock"
)

// ShellRunnerI is an autogenerated mock type for the ShellRunnerI type
type ShellRunnerI struct {
	mock.Mock
}

// ExecGitCommand provides a mock function with given fields: args, stdoutb, erroutb, dir
func (_m *ShellRunnerI) ExecGitCommand(args []string, stdoutb *bytes.Buffer, erroutb *bytes.Buffer, dir string) (*exec.Cmd, error) {
	ret := _m.Called(args, stdoutb, erroutb, dir)

	if len(ret) == 0 {
		panic("no return value specified for ExecGitCommand")
	}

	var r0 *exec.Cmd
	var r1 error
	if rf, ok := ret.Get(0).(func([]string, *bytes.Buffer, *bytes.Buffer, string) (*exec.Cmd, error)); ok {
		return rf(args, stdoutb, erroutb, dir)
	}
	if rf, ok := ret.Get(0).(func([]string, *bytes.Buffer, *bytes.Buffer, string) *exec.Cmd); ok {
		r0 = rf(args, stdoutb, erroutb, dir)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*exec.Cmd)
		}
	}

	if rf, ok := ret.Get(1).(func([]string, *bytes.Buffer, *bytes.Buffer, string) error); ok {
		r1 = rf(args, stdoutb, erroutb, dir)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewShellRunnerI creates a new instance of ShellRunnerI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewShellRunnerI(t interface {
	mock.TestingT
	Cleanup(func())
}) *ShellRunnerI {
	mock := &ShellRunnerI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
