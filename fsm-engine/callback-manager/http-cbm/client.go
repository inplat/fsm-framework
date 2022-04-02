package http_cbm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

const (
	userAgent = "fsm-engine/http-callback-client v2.0.0"
	jsonMIME  = "application/json"
)

// prepareRequest подготавливает нагрузку для запроса
func (c *HTTPCallbackManager) prepareRequest(ctx context.Context, tx model.Tx) (*CallbackEvent, error) {
	span := opentracing.SpanFromContext(ctx)

	callbackURL, err := url.ParseRequestURI(tx.CallbackURL())
	if err != nil {
		span.SetTag("error", true).LogFields(log.Message("parse callback_url error"), log.Error(err))

		return nil, status.Errorf(codes.InvalidArgument, "parse callback_url error: %s. callback_url: %s",
			err.Error(), tx.CallbackURL())
	}

	body, err := json.Marshal(tx)
	if err != nil {
		span.SetTag("error", true).LogFields(log.Message("marshall tx info error"), log.Error(err))

		return nil, status.Errorf(codes.Internal, "marshall tx summary info error: %s. tx_id: %s",
			err.Error(), tx.ID().String())
	}

	return &CallbackEvent{
		ID:          generateEvID(1),
		TxID:        tx.ID(),
		RequestURL:  callbackURL.String(),
		RequestBody: body,
		RetryN:      1,
	}, nil
}

// sendHTTP отправляет информацию по уже заданным в CallbackEvent данным
func (c *HTTPCallbackManager) sendHTTP(ctx context.Context, cbEv *CallbackEvent) (*CallbackEvent, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "send callback")
	defer span.Finish()

	span.LogFields(log.String("uuid", cbEv.ID.String()))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cbEv.RequestURL, bytes.NewReader(cbEv.RequestBody))
	if err != nil {
		return cbEv, err
	}

	// todo добавить проброс trace_id дальше
	req.Header.Set("Content-Type", jsonMIME)
	req.Header.Set("User-Agent", userAgent)

	span.LogFields(log.String("req", string(cbEv.RequestBody)))

	cbEv.RequestTimestamp = time.Now()

	resp, err := c.c.Do(req)
	if err != nil {
		span.SetTag("error", true).LogFields(log.Error(err))

		cbEv.ResponseBody = []byte(err.Error())
		cbEv.ResponseCode = -1

		return cbEv, status.Errorf(codes.Unavailable, "network callback error: %s. callback_url: %s",
			err.Error(), cbEv.RequestURL)
	}

	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	cbEv.ResponseBody = respBody
	cbEv.ResponseCode = resp.StatusCode

	if resp.StatusCode >= 300 {
		span.SetTag("error", true).
			LogFields(
				log.String("status", resp.Status),
				log.String("resp", string(respBody)))

		return cbEv, status.Errorf(codes.Unavailable, "callback status: %s. code: %d",
			resp.Status, resp.StatusCode)
	}

	cbEv.SentSuccessfully = true

	return cbEv, nil
}
