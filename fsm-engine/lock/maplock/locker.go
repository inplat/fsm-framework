package maplock

import (
	"context"
	"errors"
	"sync"

	"fsm-framework/fsm-engine/lock"
)

var (
	ErrNotObtained = errors.New("maplock: not obtained")
)

type Locker struct {
	m sync.Map
}

func NewLocker() (lock.Locker, error) {
	return &Locker{}, nil
}

func (l *Locker) Close() error {
	return nil
}

func (l *Locker) ObtainLock(ctx context.Context, key string) (lock.Lock, error) {
	if _, ok := l.m.LoadOrStore(key, true); ok {
		return nil, ErrNotObtained
	}

	return wrapLock(ctx, key, &l.m)
}

func (l *Locker) IsErrNotObtained(err error) bool {
	return errors.Is(err, ErrNotObtained)
}
