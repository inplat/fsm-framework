package test_model

import (
	"context"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

func (f *FooStateDeclaration) EventHandler(ctx context.Context, ev *model.Event) model.State {
	panic("implement me")
}
