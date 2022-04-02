package fsmengine

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
	"github.com/inplat/fsm-framework.git/misk/caller"
	zlog "github.com/inplat/fsm-framework.git/misk/logger"
)

// resolveNextState должна выяснить какое состояние должно стать следующим
func (p *processPipeline) resolveNextState(pCtx context.Context) (context.Context, error) {
	var (
		span opentracing.Span
	)

	if p.cfg.verboseTracing {
		span, _ = opentracing.StartSpanFromContext(pCtx, caller.CurrentFuncNameClear())
		defer span.Finish()
	}

	// если паники не было, состояние было успешно получено обработчика состояния
	if !p.isPanicRecovered {
		p.event.Status = model.EventStatusDone

		return pCtx, nil
	}

	// в остальных случаях просто повторяем снова
	p.event.Status = model.EventStatusRetry
	p.nextState = p.state

	return pCtx, nil
}

// nextEvent создает следующее событие, если необходимо, иначе помечает транзакцию как завершенную
func (p *processPipeline) nextEvent(pCtx context.Context) (context.Context, error) {
	var (
		span opentracing.Span
	)

	ctx := pCtx
	if p.cfg.verboseTracing {
		span, ctx = opentracing.StartSpanFromContext(pCtx, caller.CurrentFuncNameClear())
		defer span.Finish()
	}

	if p.nextState == nil {
		p.event.Tx.SetStatus(model.TxStatusDone)

		return pCtx, nil
	}

	p.event.FinalState = p.nextState.Name()
	p.event.Tx.SetState(p.nextState)
	p.event.Tx.SetStatus(model.TxStatusPending)

	// отсчет с 0
	var retryN int
	if p.nextState == p.state {
		retryN = p.event.RetryN + 1
	}

	// если это была последняя попытка
	if retryN >= model.EventRetryMaxCount-1 {
		p.event.Status = model.EventStatusError
		p.event.Tx.SetStatus(model.TxStatusError)

		if fallbackState := p.state.FallbackState(); fallbackState != nil {
			p.nextState = fallbackState
		} else {
			p.nextState = nil
		}

		zlog.Ctx(ctx).Error().Msg("max retry count exceeded")
		p.span.SetTag("error", true).LogFields(log.Message("max retry count exceeded"))
	}

	nextEvent := model.NewEvent(p.nextState, p.event.Tx, retryN)

	p.span.LogFields(
		log.String("next_state", p.nextState.Name()),
		log.String("next_event_id", nextEvent.ID.String()),
	)

	// кодируем следующее событие
	var body []byte

	body, err := model.EventMarshal(nextEvent)
	if err != nil {
		p.span.SetTag("error", true).
			LogFields(log.Message("can't marshal next event"), log.Error(err))
		zlog.Ctx(ctx).Error().Err(err).Msg("can't marshal next event")

		p.delivery.Reject(ctx)

		return pCtx, fmt.Errorf("marshal next event: %w", err)
	}

	// сохраняем его в БД
	err = p.cfg.repo.UpdateEvent(ctx, nextEvent)
	if err != nil {
		p.span.SetTag("error", true).
			LogFields(log.Message("next event update error"), log.Error(err))
		zlog.Ctx(ctx).Error().Err(err).Msg("next event update error")

		p.delivery.Reject(ctx)

		return pCtx, fmt.Errorf("next event update: %w", err)
	}

	p.nextStateMessage = body

	return pCtx, nil
}

// updateCompletedEvent обновляет информацию по текущему событию
func (p *processPipeline) updateCompletedEvent(pCtx context.Context) (context.Context, error) {
	var (
		span opentracing.Span
	)

	ctx := pCtx
	if p.cfg.verboseTracing {
		span, ctx = opentracing.StartSpanFromContext(pCtx, caller.CurrentFuncNameClear())
		defer span.Finish()
	}

	p.span.LogFields(log.String("event_status", p.event.Status.String()))

	// обновляем событие в БД
	err := p.cfg.repo.UpdateEvent(ctx, p.event)
	if err != nil {
		p.span.SetTag("error", true).
			LogFields(log.Message("event update error"), log.Error(err))
		zlog.Ctx(ctx).Warn().Err(err).Msg("event update error")
	}

	return pCtx, nil
}

// updateCompletedTx обновляет информацию о транзакции
func (p *processPipeline) updateCompletedTx(pCtx context.Context) (context.Context, error) {
	var (
		span opentracing.Span
	)

	ctx := pCtx
	if p.cfg.verboseTracing {
		span, ctx = opentracing.StartSpanFromContext(pCtx, caller.CurrentFuncNameClear())
		defer span.Finish()
	}

	err := p.cfg.repo.UpdateTransaction(ctx, p.event.Tx, p.state.Name())
	if err != nil {
		p.span.SetTag("error", true).
			LogFields(log.String("msg", "can't update transaction status"), log.Error(err))
		zlog.Ctx(ctx).Error().Err(err).Msg("can't update transaction status")

		p.delivery.Reject(ctx)

		return pCtx, fmt.Errorf("update tx: %w", err)
	}

	return pCtx, nil
}

func (p *processPipeline) sendCallback(pCtx context.Context) (context.Context, error) {
	var (
		span opentracing.Span
	)

	// skip intermediate-states (only success or fail)
	if !(p.state.IsSuccessFinal() || p.state.IsFailFinal()) {
		return pCtx, nil
	}

	ctx := pCtx
	if p.cfg.verboseTracing {
		span, ctx = opentracing.StartSpanFromContext(pCtx, caller.CurrentFuncNameClear())
		defer span.Finish()
	}

	p.cfg.cm.Send(ctx, p.event.Tx)

	return pCtx, nil
}
