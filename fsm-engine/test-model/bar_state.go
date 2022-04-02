package test_model

import (
	"time"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

var BarState model.State = &BarStateDeclaration{}

type BarStateDeclaration struct {
}

func (f *BarStateDeclaration) Name() string {
	return "TEST_TX_BAR_STATE"
}

func (f *BarStateDeclaration) EventType() string {
	return "test_tx_bar_state_event"
}

func (f *BarStateDeclaration) Queue() string {
	return "test_tx_bar_state_event_queue"
}

func (f *BarStateDeclaration) CanTransitIn(state model.State) bool {
	if state == BarState { // nolint: gosimple
		return true
	}

	return false
}

func (f *BarStateDeclaration) MaxRetiesCount() int {
	return 15
}

func (f *BarStateDeclaration) MinRetiesDelay() time.Duration {
	return 15 * time.Second
}

func (f *BarStateDeclaration) CancellationTTL() time.Duration {
	return 15 * time.Minute
}

func (f *BarStateDeclaration) FallbackState() model.State {
	return nil
}

func (f *BarStateDeclaration) IsInitial() bool {
	return false
}

func (f *BarStateDeclaration) IsSuccessFinal() bool {
	return true
}

func (f *BarStateDeclaration) IsFailFinal() bool {
	return false
}

func (f *BarStateDeclaration) Model() model.Model {
	return Model
}

func (f *BarStateDeclaration) Service() Service {
	return Model.Service().(Service)
}
