package callback_manager

import (
	"context"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

type CallbackManager interface {
	// Stop останавливает обработку новых колбеков
	Stop() error
	// Send гарантирует отправку колбека по стратегии
	Send(ctx context.Context, tx model.Tx)
}
