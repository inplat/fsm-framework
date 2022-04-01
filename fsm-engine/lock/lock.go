package lock

import (
	"context"
)

type Lock interface {
	Release(ctx context.Context) error
}

type Locker interface {
	ObtainLock(ctx context.Context, key string) (Lock, error)
	IsErrNotObtained(error) bool
	Close() error
}
