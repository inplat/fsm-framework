//go:generate mockery --name (.*)Mock

package genmocks

import (
	test_model "fsm-framework/fsm-engine/test-model"

	"fsm-framework/fsm-engine/queue"

	callback_manager "fsm-framework/fsm-engine/callback-manager"
	"fsm-framework/fsm-engine/lock"
	"fsm-framework/fsm-engine/model"
)

type QueueBrokerMock interface {
	queue.Broker
}

type QueueChannelMock interface {
	queue.Channel
}

type QueueDeliveryMock interface {
	queue.Delivery
}

type LockLockerMock interface {
	lock.Locker
}

type LockLockMock interface {
	lock.Lock
}

type TestServiceMock interface {
	test_model.Service
}

type RepositoryMock interface {
	model.Repository
}

type CallbackManagerMock interface {
	callback_manager.CallbackManager
}
