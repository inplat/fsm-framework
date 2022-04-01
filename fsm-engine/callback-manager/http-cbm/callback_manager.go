package http_cbm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	callback_manager "fsm-framework/fsm-engine/callback-manager"
	"fsm-framework/fsm-engine/model"
	"fsm-framework/fsm-engine/queue"
	zlog "fsm-framework/misk/logger"
)

const (
	// syncRetryN количество синхронных повторов отправить обратный вызов
	syncRetryN = 3
	// syncRetryDelay минимальная задержка между синхронными вызовами
	syncRetryDelay = 3 * time.Second
	// asyncRetryN количество асинхронных повторов отправить обратный вызов
	asyncRetryN = 7
	// asyncRetryDelay минимальная задержка между асинхронными вызовами
	asyncRetryDelay = 10 * time.Second
	// timeout общий таймаут ожидания ответа
	timeout = 10 * time.Second
	// queueName очередь в системе очередей
	queueName = "callback_manager"
)

var _ callback_manager.CallbackManager = &HTTPCallbackManager{}

// HTTPCallbackManager
// управляет отправкой http-колбеков,
// гарантируя доставку согласно заданным SLA
// использует систему очередей и постоянное хранилище для обеспечения доставки
type HTTPCallbackManager struct {
	r      Repository
	c      http.Client
	b      queue.Broker
	pushCh queue.Channel
	pullCh queue.Channel
}

func New(ctx context.Context, r Repository, b queue.Broker) (*HTTPCallbackManager, error) {
	var err error

	c := &HTTPCallbackManager{
		c: http.Client{
			Timeout: timeout,
		},
		r: r,
		b: b,
	}

	c.pushCh, err = c.b.Channel()
	if err != nil {
		return nil, fmt.Errorf("can't init channel for delivery producing: %w", err)
	}

	err = c.pushCh.DeclareQueue(queueName)
	if err != nil {
		return nil, fmt.Errorf("can't init queue for delivery producing: %w", err)
	}

	err = c.queueProcessing(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't init queue consuming: %w", err)
	}

	return c, nil
}

// Send реализует логику по отправке колбека с sync и async повторами
func (c *HTTPCallbackManager) Send(ctx context.Context, tx model.Tx) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "fsm send callback")
	defer span.Finish()

	// check if url not empty
	if tx.CallbackURL() == "" {
		return
	}

	// generate json
	cbEv, err := c.prepareRequest(ctx, tx)
	if err != nil {
		zlog.Ctx(ctx).Error().Err(err).Msg("can't prepare callback request")

		return
	}

	// trying to send sync
	for i := 1; i <= syncRetryN; i++ {
		if updErr := c.r.UpdateCallbackEvent(ctx, cbEv); updErr != nil {
			zlog.Ctx(ctx).Error().Err(updErr).Msg("can't update callback event before send")
		}

		cbEv, _ = c.sendHTTP(ctx, cbEv)

		if updErr := c.r.UpdateCallbackEvent(ctx, cbEv); updErr != nil {
			zlog.Ctx(ctx).Error().Err(updErr).Msg("can't update callback event after send")
		}

		if cbEv.SentSuccessfully {
			return
		}

		cbEv = cbEv.NewRetry()

		time.Sleep(syncRetryDelay)
	}

	// if not succeeded add in db
	if updErr := c.r.UpdateCallbackEvent(ctx, cbEv); updErr != nil {
		zlog.Ctx(ctx).Error().Err(updErr).Msg("can't update callback event before push to queue")
	}

	// and send to queue
	c.pushToQueue(ctx, cbEv)
}

// Stop останавливает работу с очередью
func (c *HTTPCallbackManager) Stop() error {
	err := c.pullCh.Close()
	if err != nil {
		return fmt.Errorf("error while stopping callbacks conumser: %w", err)
	}

	return nil
}
