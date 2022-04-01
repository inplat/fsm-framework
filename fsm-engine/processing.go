package fsmengine

import (
	"context"
	"runtime/debug"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	callback_manager "fsm-framework/fsm-engine/callback-manager"
	"fsm-framework/fsm-engine/lock"
	"fsm-framework/fsm-engine/model"
	"fsm-framework/fsm-engine/queue"
	zlog "fsm-framework/misk/logger"
)

type pipelineConfig struct {
	// repo репозиторий для обновления информации о транзакции и событиях
	repo model.Repository
	// cm отправка обратных уведомлений
	cm callback_manager.CallbackManager
	// ch канал для отправки в другие очереди сообщений
	ch queue.Channel
	// locker объект для получения блокировок на обработку транзакций
	locker lock.Locker
	// verboseTracing подробное логгирование в трейсинг
	verboseTracing bool
}

// processPipeline структура проводящая процесс пре/постобработки конкретного полученного из очереди сообщения
type processPipeline struct {
	// cfg зависимости для обработки транзакции
	cfg *pipelineConfig

	// state текущее состояние
	state model.State
	// delivery исходная посылка, полученная из очереди
	delivery queue.Delivery
	// txLock эксклюзивный лок на обработку транзакции
	txLock lock.Lock
	// span спан в системе трейсинга
	span opentracing.Span
	// event текущее событие
	event *model.Event

	// isPanicRecovered во время процессинга была отловлена паника
	isPanicRecovered bool
	// nextState состояние в которое планируется осуществить переход (может быть nil)
	nextState model.State
	// nextStateMessage сообщение, которое нужно отправить в очередь (может быть nil)
	nextStateMessage []byte
}

func newProcessPipeline(cfg *pipelineConfig, state model.State) *processPipeline {
	return &processPipeline{
		cfg:   cfg,
		state: state,
	}
}

// Process запускает обработку конкретного события, полученного из очереди в рамках конвейера
func (p *processPipeline) Process(ctx context.Context, delivery queue.Delivery) {
	defer p.stop(ctx)

	p.delivery = delivery

	zlog.Ctx(ctx).Trace().
		Interface("body", string(p.delivery.GetBody())).
		Msg("event received")

	var err error

	ctx, err = p.eventUnmarshal(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("event processing")

	ctx, err = p.checkRetryDelay(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("tx event was delayed if needed")

	ctx, err = p.obtainLock(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("tx lock obtained successfully")

	ctx, err = p.checkTx(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("tx state validated")

	ctx, err = p.startTracing(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("tracing started")

	ctx, err = p.updateProgressEvent(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("event updated successfully")

	ctx, err = p.updateProgressTx(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("tx updated successfully")

	zlog.Ctx(ctx).Debug().Msg("event preprocessed successfully")

	p.process(ctx)

	zlog.Ctx(ctx).Debug().Msg("event processed")

	ctx, err = p.resolveNextState(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("resolved next state successfully")

	ctx, err = p.nextEvent(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("next event creation completed")

	ctx, err = p.updateCompletedEvent(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("event updated")

	ctx, err = p.updateCompletedTx(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("tx updated")

	p.delivery.Ack(ctx)

	zlog.Ctx(ctx).Trace().Msg("queue delivery acknowledged")

	// уведомления при необходимости
	ctx, err = p.sendCallback(ctx)
	if err != nil {
		return
	}

	zlog.Ctx(ctx).Trace().Msg("tx status callback")

	zlog.Ctx(ctx).Debug().Msg("event postprocessed successfully")
}

// stop освобождает ресурсы после обработки, в том числе shared блокировки
func (p *processPipeline) stop(ctx context.Context) {
	// release lock if exists
	if p.txLock != nil {
		err := p.txLock.Release(ctx)
		if err != nil {
			zlog.Ctx(ctx).Error().Err(err).Msg("event lock can't be released")
		} else {
			zlog.Ctx(ctx).Trace().Msg("event lock released")
		}
	}

	// send queue message if exists
	if len(p.nextStateMessage) > 0 && p.nextState != nil {
		// отправляем в соответствующую очередь
		p.cfg.ch.Publish(ctx, p.nextState.Queue(), p.nextStateMessage)
	}

	// stop span
	if p.span != nil {
		p.span.Finish()
	}
}

func (p *processPipeline) process(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			p.isPanicRecovered = true

			stack := string(debug.Stack())

			zlog.Ctx(ctx).Error().
				Interface("panic", r).
				Str("stack", stack).
				Msg("panic recovered in state consumer")
			p.span.SetTag("error", true).LogFields(log.Object("panic", r), log.String("stack", stack))
		}
	}()

	sp, handlerCtx := opentracing.StartSpanFromContext(ctx, "Event Handler")
	defer sp.Finish()

	p.nextState = p.state.EventHandler(handlerCtx, p.event)
	if p.nextState != nil {
		p.span.LogFields(log.String("next_state", p.nextState.Name()))
	}
}
