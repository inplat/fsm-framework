// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// LockLockMock is an autogenerated mock type for the LockLockMock type
type LockLockMock struct {
	mock.Mock
}

// Release provides a mock function with given fields: ctx
func (_m *LockLockMock) Release(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
