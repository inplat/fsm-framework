package test_model

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inplat/fsm-framework.git/fsm-engine/genmocks/mocks"

	fsmengine "github.com/inplat/fsm-framework.git/fsm-engine"
)

// todo: разделить на отдельные тесты fsm движка

// Тестирует основные узлы связки FSM Модель + FSM Движок снаружи
func TestInit(t *testing.T) {
	var err error
	ctx := context.TODO()

	tx := &testTx{
		TxID: uuid.New(),
	}

	// init
	locker := &mocks.LockLockerMock{}
	locker.On("ObtainLock", mock.Anything, mock.Anything).
		Return(&mocks.LockLockMock{}, nil)

	repo := &mocks.RepositoryMock{}
	repo.On("CreateTransaction", mock.Anything, mock.Anything).
		Return(nil)
	repo.On("UpdateTransaction", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	repo.On("Transaction", mock.Anything, mock.Anything).
		Return(tx, nil)

	svc := &mocks.TestServiceMock{}

	qChan := &mocks.QueueChannelMock{}
	qChan.On("Consume", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	qChan.On("Publish", mock.Anything, mock.Anything, mock.Anything).
		Return()

	broker := &mocks.QueueBrokerMock{}
	broker.On("Channel").
		Return(qChan, nil)

	cm := &mocks.CallbackManagerMock{}
	cm.On("Send", mock.Anything, mock.Anything, mock.Anything).
		Return()

	engine := fsmengine.New(fsmengine.Config{
		Repository:      repo,
		Locker:          locker,
		Broker:          broker,
		CallbackManager: cm,
		VerboseTracing:  false,
	})

	// add model
	err = Model.SetService(svc)
	assert.NoError(t, err, "model set service error")
	err = engine.AddModel(ctx, Model)
	assert.NoError(t, err, "engine add model error")

	// resolving
	foundS, foundM := engine.Resolve(ctx, FooState.Name())
	assert.Equal(t, FooState, foundS, "engine state resolving error")
	assert.Equal(t, Model, foundM, "engine model resolving error")

	// new tx
	assert.NoError(t, engine.CreateTx(ctx, tx, FooState))
	assert.Error(t, engine.CreateTx(ctx, tx, BarState))

	// transit
	assert.NoError(t, engine.Transit(ctx, tx, BarState))
}
