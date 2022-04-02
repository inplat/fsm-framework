package test_model

import (
	"time"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

var FooState model.State = &FooStateDeclaration{}

type FooStateDeclaration struct {
}

func (f *FooStateDeclaration) Name() string {
	return "TEST_TX_FOO_STATE"
}

func (f *FooStateDeclaration) EventType() string {
	return "test_tx_foo_state_event"
}

func (f *FooStateDeclaration) Queue() string {
	return "test_tx_foo_state_event_queue"
}

func (f *FooStateDeclaration) CanTransitIn(state model.State) bool {
	if state == f {
		return true
	}

	if state == BarState {
		return true
	}

	return false
}

func (f *FooStateDeclaration) MaxRetiesCount() int {
	return 15
}

func (f *FooStateDeclaration) MinRetiesDelay() time.Duration {
	return 15 * time.Second
}

func (f *FooStateDeclaration) CancellationTTL() time.Duration {
	return 15 * time.Minute
}

func (f *FooStateDeclaration) FallbackState() model.State {
	return BarState
}

func (f *FooStateDeclaration) IsInitial() bool {
	return true
}

func (f *FooStateDeclaration) IsSuccessFinal() bool {
	return false
}

func (f *FooStateDeclaration) IsFailFinal() bool {
	return false
}

func (f *FooStateDeclaration) Model() model.Model {
	return Model
}

func (f *FooStateDeclaration) Service() Service {
	return Model.Service().(Service)
}
