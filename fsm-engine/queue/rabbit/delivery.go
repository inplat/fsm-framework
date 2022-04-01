package rabbit

import (
	"context"

	"github.com/streadway/amqp"

	zlog "fsm-framework/misk/logger"
)

type Delivery struct {
	d amqp.Delivery
}

func (d *Delivery) Ack(ctx context.Context) {
	err := d.d.Ack(false)
	if err == nil {
		return
	}

	// todo: retry
	zlog.Ctx(ctx).Error().Err(err).Msg("rabbitmq ack error")
}

func (d *Delivery) Reject(ctx context.Context) {
	err := d.d.Nack(false, true)
	if err == nil {
		return
	}

	// todo: retry
	zlog.Ctx(ctx).Error().Err(err).Msg("rabbitmq reject error")
}

func (d *Delivery) GetBody() []byte {
	return d.d.Body
}
