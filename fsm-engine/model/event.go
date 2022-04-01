package model

import (
	"encoding/json"
	"fmt"
	"time"

	"fsm-framework/misk/prettyuuid"

	"github.com/google/uuid"
)

const (
	// EventRetryMaxCount максимальное количество попыток повтора обработки события
	EventRetryMaxCount = 5
	// EventRetryMinDelay минимально необходимая задержка между обработкой одного и того же сообщения, если лок занят
	EventRetryMinDelay = 15 * time.Second
)

type EventStatus string

func (e EventStatus) String() string {
	return string(e)
}

const (
	// EventStatusPending событие ожидает своей обработки в очереди
	EventStatusPending EventStatus = "pending"
	// EventStatusProgress событие обрабатывается консюмером прямо сейчас
	EventStatusProgress EventStatus = "progress"
	// EventStatusDone событие было обработано успешно
	EventStatusDone EventStatus = "done"
	// EventStatusRetry ошибка при обработке события, попытка повторить обработку позже другим костюмером
	EventStatusRetry EventStatus = "retry"
	// EventStatusError выставляется этот статус в случае, если ошибка в обработке события
	// повторилась максимальное кол-вол раз, значит, мы не можем никуда перейти из этого состояния
	EventStatusError EventStatus = "error"
)

// Event (событие) – структура являющаяся сообщением, передаваемым в очереди.
// Исполнение события (а точнее его обработчика) меняет состояние транзакции.
// События не могут быть переиспользованы, вместо этого создается новое событие,
// которое принимает от родителя транзакцию (Tx) и номер повторения (RetryN), отчет ведется с 0.
type Event struct {
	Tx         Tx          `json:"tx"`
	ID         uuid.UUID   `json:"event_id" db:"event_id"`
	Type       string      `json:"event_type" db:"event_type"`
	StartState string      `json:"start_state" db:"start_state"`
	FinalState string      `json:"final_state" db:"final_state"`
	Status     EventStatus `json:"status" db:"status"`
	// RetryN попытки, отсчет с 0
	RetryN  int       `json:"retry_n" db:"retry_n"`
	SpanID  string    `json:"span_id" db:"span_id"`
	Updated time.Time `json:"updated" db:"updated"`
	Created time.Time `json:"created" db:"created"`
}

func EventMarshal(e *Event) ([]byte, error) {
	return json.Marshal(e)
}

func EventUnmarshal(body []byte) (*Event, error) {
	ev := &Event{}

	err := json.Unmarshal(body, ev)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshall event: %w", err)
	}

	return ev, nil
}

func NewEvent(state State, tx Tx, retryN int) *Event {
	// creates ID in form of EE00xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx for easy debugging
	UUID := prettyuuid.New(0xEE, 0x00)

	return &Event{
		ID:         UUID,
		Tx:         tx,
		Type:       state.EventType(),
		StartState: state.Name(),
		FinalState: "",
		Status:     EventStatusPending,
		RetryN:     retryN,
		SpanID:     "",
		Updated:    time.Now(),
		Created:    time.Now(),
	}
}
