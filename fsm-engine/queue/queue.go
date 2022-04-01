package queue

import (
	"context"
)

type Delivery interface {
	// Ack подтверждение принятия сообщения, внутри сокрыта логика ретраев и логгирования ошибки
	Ack(ctx context.Context)
	// Reject сообщение не удалось обработать, внутри сокрыта логика ретраев и логгирования ошибки
	// todo requeue param
	Reject(ctx context.Context)
	// GetBody тело сообщения
	GetBody() []byte
}

// Handler функция обработчик посылки, полученной из очереди
type Handler func(ctx context.Context, d Delivery) error

type Channel interface {
	// Close закрывает канал, ожидая завершения всех хендлеров текущих сообщений
	Close() error
	// DeclareQueue создает очередь, если она не создана, либо возвращает ошибку
	DeclareQueue(name string) error
	// Publish кладет сообщение в очередь, внутри сокрыта логика ретраев и логгирования ошибки
	Publish(ctx context.Context, queue string, body []byte)
	// Consume запускает обработку входящих сообщений (async), ошибка может быть только при создании
	Consume(ctx context.Context, queue string, handler Handler) error
}

type Broker interface {
	// Close закрывает все соединения с очередью
	Close(ctx context.Context)
	// Channel создает канал очереди, внутри сокрыта балансировка между подключениями к брокеру
	Channel() (Channel, error)
}
