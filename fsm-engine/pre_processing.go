package fsmengine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"

	"fsm-framework/fsm-engine/model"
	"fsm-framework/misk/caller"
	zlog "fsm-framework/misk/logger"
)

const (
	lockPrefix = "lock_tx:"
)

// eventUnmarshal декодирует полученное из очереди сообщение в model.Event, проверяет, что tx_id не пустое
func (p *processPipeline) eventUnmarshal(ctx context.Context) (context.Context, error) {
	var err error

	// json decode
	p.event, err = model.EventUnmarshal(p.delivery.GetBody())
	if err != nil {
		zlog.Ctx(ctx).Error().Err(err).Msg("consumer message unmarshal error")

		p.delivery.Ack(ctx)

		return ctx, fmt.Errorf("event unmarshall: %w", err)
	}

	// nullable check
	if p.event.Tx.ID() == uuid.Nil {
		zlog.Ctx(ctx).Error().Str("event_id", p.event.ID.String()).Msg("consumer message tx_id is nil")

		p.delivery.Ack(ctx)

		return ctx, errors.New("null tx_id")
	}

	// log prepend
	ctx = zlog.FromLogger(zlog.Ctx(ctx).With().
		Str("event_id", p.event.ID.String()).
		Str("tx_id", p.event.Tx.ID().String()).Logger()).WithContext(ctx)

	return ctx, nil
}

// checkRetryDelay спит, если с момента создания события прошло меньше model.EventRetryMinDelay
func (p *processPipeline) checkRetryDelay(pCtx context.Context) (context.Context, error) {
	if p.event.RetryN == 0 {
		return pCtx, nil
	}

	since := time.Since(p.event.Created)
	diff := model.EventRetryMinDelay - since

	if diff <= 0 {
		return pCtx, nil
	}

	zlog.Ctx(pCtx).Trace().Dur("duration", diff).Msg("sleep for delaying event processing")
	time.Sleep(diff)

	return pCtx, nil
}

// obtainLock попытка заполучить эксклюзивную блокировку на обработку события
func (p *processPipeline) obtainLock(ctx context.Context) (context.Context, error) {
	var err error

	p.txLock, err = p.cfg.locker.ObtainLock(ctx, lockPrefix+p.event.Tx.ID().String())
	if err != nil {
		if p.cfg.locker.IsErrNotObtained(err) {
			zlog.Ctx(ctx).Debug().Err(err).Msg("lock already obtained by other consumer")
			time.Sleep(model.EventRetryMinDelay) // жесткий фикс ретраев для локов
		} else {
			err = fmt.Errorf("lock obtaining error: %w", err)
			zlog.Ctx(ctx).Error().Err(err).Msg("consumer lock obtain failed")
		}

		p.delivery.Reject(ctx)

		return ctx, err
	}

	return ctx, nil
}

// checkTx проверяет состояние транзакции, находится ли она в БД вообще
func (p *processPipeline) checkTx(ctx context.Context) (context.Context, error) {
	tx, err := p.cfg.repo.Transaction(ctx, p.event.Tx.ID())
	if err != nil {
		zlog.Ctx(ctx).Error().Err(err).Msg("can't get transaction from db")

		p.delivery.Ack(ctx)

		return ctx, fmt.Errorf("transaction repository get: %w", err)
	}

	if p.event.Tx.ID() != tx.ID() || tx.State() != p.state {
		zlog.Ctx(ctx).Error().Str("current_state", tx.State().Name()).Msg("transaction state incompatible")

		p.delivery.Ack(ctx)

		return ctx, fmt.Errorf("transaction state incompatible")
	}

	// актуализируем информацию
	p.event.Tx = tx

	return ctx, nil
}

// startTracing создает новый span для текущей обработки события
func (p *processPipeline) startTracing(ctx context.Context) (context.Context, error) {
	var parentSpanCtx opentracing.SpanContext

	traceID, traceErr := jaeger.TraceIDFromString(p.event.Tx.TraceID())
	parentSpanID, spanErr := jaeger.SpanIDFromString(p.event.Tx.SpanID())

	if traceErr == nil && spanErr == nil {
		parentSpanCtx = jaeger.NewSpanContext(traceID, parentSpanID, parentSpanID, true, nil)
	}

	// Либо наследуемся ChildOf, либо будет просто отдельная ветка спанов
	p.span = opentracing.StartSpan(
		"Event: "+p.state.Name(),
		opentracing.ChildOf(parentSpanCtx),
	)
	// p.span.Finish() выполняется в методе processPipeline.stop

	p.span.LogFields(
		log.String("event_id", p.event.ID.String()),
		log.String("tx_id", p.event.Tx.ID().String()),
		log.Int("retry_n", p.event.RetryN),
	)

	ctx = opentracing.ContextWithSpan(ctx, p.span)

	return ctx, nil
}

// updateProgressEvent обновляет событие в БД
func (p *processPipeline) updateProgressEvent(pCtx context.Context) (context.Context, error) {
	var (
		span opentracing.Span
	)

	ctx := pCtx
	if p.cfg.verboseTracing {
		span, ctx = opentracing.StartSpanFromContext(pCtx, caller.CurrentFuncNameClear())
		defer span.Finish()
	}

	p.event.Status = model.EventStatusProgress
	if jSpan, ok := p.span.(*jaeger.Span); ok {
		p.event.SpanID = jSpan.SpanContext().SpanID().String()
	} else {
		zlog.Ctx(ctx).Warn().Msg("can't eject trace_id from root span")
	}

	err := p.cfg.repo.UpdateEvent(ctx, p.event)
	if err != nil {
		p.span.LogFields(log.Message("event update error"), log.Error(err))
		zlog.Ctx(ctx).Warn().Err(err).Msg("event update error")
	}

	return pCtx, nil
}

// updateProgressTx обновляет транзакцию в БД, устанавливает ей новый статус
func (p *processPipeline) updateProgressTx(pCtx context.Context) (context.Context, error) {
	var (
		span opentracing.Span
	)

	ctx := pCtx
	if p.cfg.verboseTracing {
		span, ctx = opentracing.StartSpanFromContext(pCtx, caller.CurrentFuncNameClear())
		defer span.Finish()
	}

	p.event.Tx.SetStatus(model.TxStatusProgress)

	err := p.cfg.repo.UpdateTransaction(ctx, p.event.Tx, p.event.Tx.State().Name())
	if err != nil {
		p.span.SetTag("error", true).
			LogFields(log.Message("tx update error"), log.Error(err))
		zlog.Ctx(ctx).Error().Err(err).Msg("tx status update error")

		p.delivery.Reject(ctx)

		return pCtx, fmt.Errorf("tx status update error: %w", err)
	}

	return pCtx, nil
}
