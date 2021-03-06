// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	queue "fsm-framework/fsm-engine/queue"
)

// QueueBrokerMock is an autogenerated mock type for the QueueBrokerMock type
type QueueBrokerMock struct {
	mock.Mock
}

// Channel provides a mock function with given fields:
func (_m *QueueBrokerMock) Channel() (queue.Channel, error) {
	ret := _m.Called()

	var r0 queue.Channel
	if rf, ok := ret.Get(0).(func() queue.Channel); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Channel)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Close provides a mock function with given fields: ctx
func (_m *QueueBrokerMock) Close(ctx context.Context) {
	_m.Called(ctx)
}
