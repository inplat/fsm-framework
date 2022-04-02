package test_model

import (
	"context"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

func (f *BarStateDeclaration) EventHandler(ctx context.Context, ev *model.Event) model.State {
	// do some work with svc
	f.Service().Foo("some data")

	return FooState
}
