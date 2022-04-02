package test_model

import (
	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

var Model model.Model = &ModelDeclaration{}

type ModelDeclaration struct {
	model model.Engine
	svc   Service
}

func (m *ModelDeclaration) Name() string {
	return "test"
}

func (m *ModelDeclaration) States() []model.State {
	return []model.State{
		FooState,
		BarState,
	}
}

func (m *ModelDeclaration) Resolve(name string) model.State {
	switch name {
	case FooState.Name():
		return FooState
	case BarState.Name():
		return BarState
	}

	return nil
}

func (m *ModelDeclaration) Has(state model.State) bool {
	if state == FooState {
		return true
	}

	if state == BarState {
		return true
	}

	return false
}

func (m *ModelDeclaration) SetEngine(engine model.Engine) {
	m.model = engine
}

func (m *ModelDeclaration) Engine() model.Engine {
	return m.model
}

func (m *ModelDeclaration) SetService(svc interface{}) error {
	var ok bool

	m.svc, ok = svc.(Service)
	if !ok {
		return model.ErrServiceNotImplementedInterface
	}

	return nil
}

func (m *ModelDeclaration) Service() interface{} {
	return m.svc
}
