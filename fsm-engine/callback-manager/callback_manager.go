package callback_manager

import (
	"context"

	"fsm-framework/fsm-engine/model"
)

type CallbackManager interface {
	// Stop останавливает обработку новых колбеков
	Stop() error
	// Send гарантирует отправку колбека по стратегии
	Send(ctx context.Context, tx model.Tx)
}
