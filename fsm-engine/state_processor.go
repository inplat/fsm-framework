package fsmengine

import (
	"context"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
	"github.com/inplat/fsm-framework.git/fsm-engine/queue"
	zlog "github.com/inplat/fsm-framework.git/misk/logger"
)

type StateProcessor struct {
	state       model.State
	consumerCh  queue.Channel
	publisherCh queue.Channel
}

func (sp *StateProcessor) StartConsume(appCtx context.Context, cfg *pipelineConfig) error {
	ctx := zlog.FromLogger(zlog.Ctx(appCtx).With().Str("state", sp.state.Name()).Logger()).
		WithContext(context.Background())

	err := sp.consumerCh.Consume(appCtx, sp.state.Queue(), func(_ context.Context, d queue.Delivery) error {
		newProcessPipeline(cfg, sp.state).Process(ctx, d)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (sp *StateProcessor) Publish(ctx context.Context, ev *model.Event) {
	body, err := model.EventMarshal(ev)
	if err != nil {
		zlog.Ctx(ctx).Error().
			Interface("event", ev).
			Err(err).
			Msg("event marshalling error")

		return
	}

	sp.publisherCh.Publish(ctx, sp.state.Queue(), body)
}
