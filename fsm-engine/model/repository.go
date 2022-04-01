package model

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Transaction вычитывает данные о транзакции по её ID
	Transaction(ctx context.Context, txID uuid.UUID) (Tx, error)
	// UpdateTransaction обновляет данные о транзакции, если она существует и её текущий статус совпадает с currState
	UpdateTransaction(ctx context.Context, tx Tx, currState string) error
	// CreateTransaction записывает новую транзакцию в хранилище
	CreateTransaction(ctx context.Context, tx Tx) error
	// UpdateEvent обновляет событие, создает его если еще не создано
	// опционально, сейчас используется для аудита
	UpdateEvent(ctx context.Context, event *Event) error
}
