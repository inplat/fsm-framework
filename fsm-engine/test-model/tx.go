package test_model

import (
	"github.com/google/uuid"

	"github.com/inplat/fsm-framework.git/fsm-engine/model"
)

var _ model.Tx = &testTx{}

type testTx struct {
	TxID          uuid.UUID
	TxState       model.State
	TxStatus      model.TxStatus
	TxCallbackURL string
	TxTraceID     string
	TxSpanID      string
}

func (t *testTx) ID() uuid.UUID {
	return t.TxID
}

func (t *testTx) State() model.State {
	return t.TxState
}

func (t *testTx) SetState(state model.State) {
	t.TxState = state
}

func (t *testTx) Status() model.TxStatus {
	return t.TxStatus
}

func (t *testTx) SetStatus(status model.TxStatus) {
	t.TxStatus = status
}

func (t *testTx) TraceID() string {
	return t.TxTraceID
}

func (t *testTx) SetTraceID(traceID string) {
	t.TxTraceID = traceID
}

func (t *testTx) SpanID() string {
	return t.TxSpanID
}

func (t *testTx) SetSpanID(spanID string) {
	t.TxSpanID = spanID
}

func (t *testTx) CallbackURL() string {
	return t.TxCallbackURL
}

func (t *testTx) SetCallbackURL(callbackURL string) {
	t.TxCallbackURL = callbackURL
}
