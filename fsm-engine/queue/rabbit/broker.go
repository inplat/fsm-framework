package rabbit

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"

	"github.com/inplat/fsm-framework.git/fsm-engine/queue"
	zlog "github.com/inplat/fsm-framework.git/misk/logger"
)

// channelPerConn кол-во каналов которые могут работать на одном соединении
const channelPerConn = 64

type Connection struct {
	*rabbitmq.Connection
	channels []*Channel
}

type Broker struct {
	// appName имя сервиса, чтобы прокинуть в metadata (пример: morpheus)
	appName string
	// url подключения к брокеру
	url string
	// m лок на создание нового канала
	m sync.Mutex
	// conn список подключений
	conn []*Connection
}

func NewBroker(url string, appName string) (queue.Broker, error) {
	b := &Broker{
		appName: appName,
		url:     url,
	}

	return b, nil
}

func (b *Broker) anotherConn() (*Connection, error) {
	for _, conn := range b.conn {
		if len(conn.channels) < channelPerConn {
			return conn, nil
		}
	}

	aConn, err := rabbitmq.Dial(b.url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq new conn error: %w", err)
	}

	conn := &Connection{
		Connection: aConn,
		channels:   make([]*Channel, 0, 3),
	}

	b.conn = append(b.conn, conn)

	return conn, nil
}

func (b *Broker) Channel() (queue.Channel, error) {
	b.m.Lock()
	defer b.m.Unlock()

	conn, err := b.anotherConn()
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	channel := &Channel{
		appName: b.appName,
		ch:      ch,
	}
	conn.channels = append(conn.channels, channel)

	return channel, nil
}

func (b *Broker) Close(ctx context.Context) {
	b.m.Lock()
	defer b.m.Unlock()

	for _, conn := range b.conn {
		for _, channel := range conn.channels {
			if err := channel.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
				zlog.Ctx(ctx).Err(err).Msg("error while closing broker channel")
			}
		}

		if err := conn.Close(); err != nil {
			zlog.Ctx(ctx).Err(err).Msg("error while closing broker conn")
		}
	}
}
