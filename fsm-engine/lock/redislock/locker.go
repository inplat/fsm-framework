package redislock

import (
	"context"
	"errors"
	"fmt"
	"time"

	rl "github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"

	"github.com/inplat/fsm-framework.git/fsm-engine/lock"
)

const (
	redisTimeout     = 30 * time.Second
	redisIdleTimeout = 5 * time.Minute
	redisMaxRetries  = 5
	defaultLockTTL   = 30 * time.Second // todo: поставить нормальное время для лока
)

type Locker struct {
	redisClient redis.UniversalClient
	redisLocker *rl.Client
	appName     string
}

func NewLocker(addr []string, password string, appName string) (lock.Locker, error) {
	var client redis.UniversalClient

	if len(addr) == 1 {
		opt := &redis.Options{
			Addr:         addr[0],
			Password:     password,
			MaxRetries:   redisMaxRetries,
			DialTimeout:  redisTimeout,
			ReadTimeout:  redisTimeout,
			WriteTimeout: redisTimeout,
			PoolTimeout:  redisTimeout,
			IdleTimeout:  redisIdleTimeout,
		}

		client = redis.NewClient(opt)
	} else {
		opt := &redis.ClusterOptions{
			Addrs:        addr,
			Password:     password,
			MaxRetries:   redisMaxRetries,
			DialTimeout:  redisTimeout,
			ReadTimeout:  redisTimeout,
			WriteTimeout: redisTimeout,
			PoolTimeout:  redisTimeout,
			IdleTimeout:  redisIdleTimeout,
		}

		client = redis.NewClusterClient(opt)
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("redis ping error: %w", err)
	}

	redisLocker := rl.New(client)

	return &Locker{
		redisClient: client,
		redisLocker: redisLocker,
		appName:     appName,
	}, nil
}

func (l *Locker) Close() error {
	return l.redisClient.Close()
}

func (l *Locker) ObtainLock(ctx context.Context, key string) (lock.Lock, error) {
	rlLock, err := l.redisLocker.Obtain(ctx, key, defaultLockTTL, &rl.Options{
		Metadata: l.appName,
	})
	if err != nil {
		return nil, fmt.Errorf("obtain lock err: %w", err)
	}

	return wrapLock(ctx, rlLock)
}

func (l *Locker) IsErrNotObtained(err error) bool {
	return errors.Is(err, rl.ErrNotObtained)
}
