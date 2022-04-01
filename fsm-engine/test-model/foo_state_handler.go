package test_model

import (
	"context"

	"fsm-framework/fsm-engine/model"
)

func (f *FooStateDeclaration) EventHandler(ctx context.Context, ev *model.Event) model.State {
	panic("implement me")
}
