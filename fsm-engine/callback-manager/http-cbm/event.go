package http_cbm

import (
	"time"

	"github.com/google/uuid"

	"fsm-framework/misk/prettyuuid"
)

// generateEvID создает ID по маске CABANNxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx for easy debugging
// CABA - постоянный префикс (мнем. CA(ll)BA(ck))
// NN - номер попытки в hex
func generateEvID(retryN int) uuid.UUID {
	return prettyuuid.New(0xCA, 0xBA, byte(retryN))
}

// CallbackEvent событие попытки отправить обратный вызов
type CallbackEvent struct {
	// ID уникальный идентификатор попытки отправить обратный вызов
	ID uuid.UUID `json:"id" db:"id"`
	// TxID идентификатор платежной транзакции
	TxID uuid.UUID `json:"tx_id" db:"tx_id"`
	// RequestURL адрес запроса
	RequestURL string `json:"request_url" db:"request_url"`
	// RequestBody тело запроса от сервиса
	RequestBody []byte `json:"request_body" db:"request_body"`
	// RequestTimestamp время последней отправки обратного вызова
	RequestTimestamp time.Time `json:"request_timestamp" db:"request_timestamp"`
	// ResponseCode кода ответа сервера
	ResponseCode int `json:"response_code" db:"response_code"`
	// Response ответ сервера
	ResponseBody []byte `json:"response_body" db:"response_body"`
	// RetryN порядковый номер попытки
	RetryN int `json:"retry_n" db:"retry_n"`
	// SentSuccessfully если истина, то отправлена успешно
	SentSuccessfully bool `json:"sent_successfully" db:"sent_successfully"`
}

// NewRetry создает новое событие отправки колбека с увеличенным счетчиком ретраев
func (ce *CallbackEvent) NewRetry() *CallbackEvent {
	UUID := ce.ID

	UUID[3] = byte(ce.RetryN + 1)

	return &CallbackEvent{
		ID:               UUID,
		TxID:             ce.TxID,
		RequestURL:       ce.RequestURL,
		RequestBody:      ce.RequestBody,
		RequestTimestamp: ce.RequestTimestamp,
		RetryN:           ce.RetryN + 1,
	}
}
