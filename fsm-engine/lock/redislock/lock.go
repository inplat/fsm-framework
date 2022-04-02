package redislock

import (
	"context"

	rl "github.com/bsm/redislock"

	"github.com/inplat/fsm-framework.git/fsm-engine/lock"
)

type Lock struct {
	rlLock *rl.Lock
}

// todo: metrics, opentracing
func wrapLock(ctx context.Context, rlLock *rl.Lock) (lock.Lock, error) {
	l := &Lock{
		rlLock: rlLock,
	}

	return l, nil
}

func (l *Lock) Release(ctx context.Context) error {
	return l.rlLock.Release(ctx)
}
