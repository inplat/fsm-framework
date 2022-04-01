package model

import (
	"github.com/google/uuid"
)

type TxStatus string

const (
	// TxStatusPending транзакция ожидает своей обработки в очереди
	TxStatusPending TxStatus = "pending"
	// TxStatusProgress транзакция обрабатывается прямо сейчас
	TxStatusProgress TxStatus = "progress"
	// TxStatusError транзакция застряла при обработке последнего события
	TxStatusError TxStatus = "error"
	// TxStatusDone транзакция была обработана, достигнуто терминальное состояние (в том числе Failed в FSM модели)
	TxStatusDone TxStatus = "done"
)

func (t TxStatus) String() string {
	return string(t)
}

// Tx информация о транзакции, основной источник истины: БД
// Структура также заключена в событиях (Event), которые передаются в очереди обработки
type Tx interface {
	ID() uuid.UUID

	State() State
	SetState(State)

	Status() TxStatus
	SetStatus(TxStatus)

	TraceID() string
	SetTraceID(traceID string)
	SpanID() string
	SetSpanID(spanID string)

	CallbackURL() string
	SetCallbackURL(callbackURL string)
}
