//go:generate mockery --name (.*)Mock

package genmocks

import (
	test_model "github.com/inplat/fsm-framework.git/fsm-engine/test-model"

	"github.com/inplat/fsm-framework.git/fsm-engine/queue"

	callback_manager "github.com/inplat/fsm-framework.git/fsm-engine/callback-manager"
	"github.com/inplat/fsm-framework.git/fsm-engine/lock"
	"github.com/inplat/fsm-framework.git/fsm-engine/model"
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
