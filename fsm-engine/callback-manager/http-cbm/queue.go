package http_cbm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/inplat/fsm-framework.git/fsm-engine/queue"
	zlog "github.com/inplat/fsm-framework.git/misk/logger"
)

// pushToQueue добавляет новую задачу отправки обратного вызова
func (c *HTTPCallbackManager) pushToQueue(ctx context.Context, cbEv *CallbackEvent) {
	eventJSON, err := json.Marshal(cbEv)
	if err != nil {
		zlog.Ctx(ctx).Error().Err(err).Msg("callback event can't be marshalled")

		return
	}

	c.pushCh.Publish(ctx, queueName, eventJSON)
}

// queueProcessing запускает обработку очереди в фоне
func (c *HTTPCallbackManager) queueProcessing(ctx context.Context) error {
	var err error

	c.pullCh, err = c.b.Channel()
	if err != nil {
		return fmt.Errorf("can't init channel for callback processing: %w", err)
	}

	err = c.pullCh.Consume(ctx, queueName, c.handler)
	if err != nil {
		return fmt.Errorf("can't run consumer for callback processing: %w", err)
	}

	return nil
}

// handler обработчик сообщения из очереди
// фактически – логика работы с async обратными вызовами
func (c *HTTPCallbackManager) handler(ctx context.Context, d queue.Delivery) error {
	cbEv := &CallbackEvent{}

	err := json.Unmarshal(d.GetBody(), &cbEv)
	if err != nil {
		zlog.Ctx(ctx).Error().Err(err).Msg("can't unmarshal callback event")

		d.Ack(ctx)

		return nil
	}

	// задержка
	time.Sleep(time.Until(cbEv.RequestTimestamp.Add(asyncRetryDelay)))

	// попытка отправить http запрос
	cbEv, _ = c.sendHTTP(ctx, cbEv)

	if updErr := c.r.UpdateCallbackEvent(ctx, cbEv); updErr != nil {
		zlog.Ctx(ctx).Error().Err(updErr).Msg("can't update callback event after send in queue")
	}

	if cbEv.SentSuccessfully {
		d.Ack(ctx)

		return nil
	}

	// если это была последняя попытка – заканчиваем обработку
	if cbEv.RetryN >= syncRetryN+asyncRetryN {
		d.Ack(ctx)

		return nil
	}

	// иначе создаем новую попытку
	cbEv = cbEv.NewRetry()

	if updErr := c.r.UpdateCallbackEvent(ctx, cbEv); updErr != nil {
		zlog.Ctx(ctx).Error().Err(updErr).Msg("can't update callback event before push to queue")
	}

	c.pushToQueue(ctx, cbEv)

	d.Ack(ctx)

	return nil
}
