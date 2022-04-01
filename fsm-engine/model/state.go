package model

import (
	"context"
	"time"
)

type State interface {
	Name() string
	EventType() string
	Queue() string
	EventHandler(ctx context.Context, ev *Event) State
	CanTransitIn(state State) bool
	MaxRetiesCount() int
	MinRetiesDelay() time.Duration
	CancellationTTL() time.Duration
	FallbackState() State
	IsInitial() bool
	IsSuccessFinal() bool
	IsFailFinal() bool
	Model() Model
}
