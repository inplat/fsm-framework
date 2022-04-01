package http_cbm

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// CallbackEvent вычитывает данные об отправке колбека по ID
	CallbackEvent(ctx context.Context, cbEventID uuid.UUID) (*CallbackEvent, error)
	// UpdateCallbackEvent записывает данные об отправке колбека по ID
	UpdateCallbackEvent(ctx context.Context, cbEvent *CallbackEvent) error
}
