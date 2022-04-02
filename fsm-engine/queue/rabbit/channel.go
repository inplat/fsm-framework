package rabbit

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
	"go.uber.org/atomic"

	"github.com/inplat/fsm-framework.git/fsm-engine/queue"
	zlog "github.com/inplat/fsm-framework.git/misk/logger"
)

var ErrConsumerExists = errors.New("consumer already exists")

type Channel struct {
	// appName имя сервиса, чтобы прокинуть в metadata (пример: morpheus)
	appName string
	// ch канал, абстракция amqp, в рамках одного соединения может быть несколько
	ch *rabbitmq.Channel
	// consuming true – канал уже используется для консюминга (канал может быть использован только для чего-то одного)
	consuming atomic.Bool
	// wg в будущем возможно будет реализована много поточная обработка сообщений из одной очереди
	wg sync.WaitGroup
}

func (c *Channel) Close() error {
	c.consuming.Store(false)
	c.wg.Wait()

	return c.ch.Close()
}

func (c *Channel) DeclareQueue(name string) error {
	_, err := c.ch.QueueDeclare(name, false, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Channel) Publish(ctx context.Context, queue string, body []byte) {
	if c.consuming.Load() {
		zlog.Ctx(ctx).Error().Msg("publish canceled, consumer-only channel")
		return
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		AppId:        c.appName,
		Body:         body,
	}

	err := c.ch.Publish("", string(queue), false, false, msg)
	if err == nil {
		return
	}

	// todo: retry
	zlog.Ctx(ctx).Error().Err(err).Msg("rabbitmq publish error")
}

func (c *Channel) consumeDelivery(ctx context.Context, h queue.Handler, d amqp.Delivery) {
	c.wg.Add(1)
	defer c.wg.Done()

	delivery := &Delivery{
		d: d,
	}

	err := h(ctx, delivery)
	if err != nil {
		delivery.Reject(ctx)
	}
}

func (c *Channel) Consume(ctx context.Context, q string, h queue.Handler) error {
	// проверяем, не занят ли канал другим консюмером
	if !c.consuming.CAS(false, true) {
		return ErrConsumerExists
	}

	err := c.DeclareQueue(q)
	if err != nil {
		return err
	}

	dCh, err := c.ch.Consume(q, "", false, false, false, false, nil)
	if err != nil {
		c.consuming.Store(false)

		return err
	}

	go func() {
		logger := zlog.FromLogger(zlog.Ctx(ctx).With().Str("queue", q).Logger())

		for d := range dCh {
			if !c.consuming.Load() {
				break
			}

			ctxWithLogger := logger.WithContext(context.Background())

			c.consumeDelivery(ctxWithLogger, h, d)
		}
	}()

	zlog.Ctx(ctx).Info().Str("queue", q).Msg("consumer started")

	return nil
}
