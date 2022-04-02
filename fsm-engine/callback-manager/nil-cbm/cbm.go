package nil_cbm

import (
	"context"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

func New() *NilCallbackManger {
	return &NilCallbackManger{}
}

type NilCallbackManger struct {
}

func (n NilCallbackManger) Stop() error {
	return nil
}

func (n NilCallbackManger) Send(ctx context.Context, tx model.Tx) {
}
