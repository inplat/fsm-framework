package maplock

import (
	"context"
	"sync"

	"fsm-framework/fsm-engine/lock"
)

type Lock struct {
	key string
	m   *sync.Map
}

func wrapLock(ctx context.Context, key string, m *sync.Map) (lock.Lock, error) {
	l := &Lock{
		key: key,
		m:   m,
	}

	return l, nil
}

func (l *Lock) Release(ctx context.Context) error {
	l.m.Delete(l.key)
	return nil
}
