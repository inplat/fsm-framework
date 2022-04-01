//go:build rnd
// +build rnd

package rabbit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testMsg = `{
  "event_id":"18d59430-7166-40e4-8a20-076b061e959b",
  "tx":{
    "tx_id":"90b1fb49-d482-4175-b3c3-97eea2c09aad",
    "state":"FIRST_STATE",
    "status":"pending",
    "amount":10260,
    "currency":"RUB",
    "merchant_id":87094309677,
    "order_id":"order-123456",
    "keep_uniq":true,
    "user_info_vk_id":90522656,
    "bonus_params_top_up":false,
    "bonus_params_spend":true,
    "bonus_params_hold":true,
    "callback_url":"https://spatecon.ru/payment/callback",
    "description":"Пополнение мобильного +79175657424",
    "created":"2021-06-29T14:24:37.175941+03:00",
    "updated":"2021-06-29T14:24:37.175941+03:00"
  },
  "event_type":"first_state_event",
  "start_state":"FIRST_STATE",
  "final_state":"",
  "status":"pending",
  "retry_n":0,
  "trace_id":"",
  "span_id":"",
  "updated":"2021-06-29T14:24:37.175944+03:00",
  "created":"2021-06-29T14:24:37.175944+03:00"
}`

func TestRabbitMQ(t *testing.T) {
	ad, err := NewBroker("amqp://localhost:5672", "morpheus-tests")
	assert.NoError(t, err)

	ch, err := ad.Channel()
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	q, err := ch.DeclareQueue("first_state_event_queue")
	assert.NoError(t, err)

	ch.Publish(ctx, q, []byte(testMsg))
	t.Log("published 1")

	deliveries, err := ch.Consume(ctx, q)
	assert.NoError(t, err)

	go func() {
		for delivery := range deliveries {
			t.Log("delivery received len=", len(delivery.GetBody()))
			delivery.Ack(ctx)
		}
	}()

	<-ctx.Done()
}
