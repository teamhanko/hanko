package flow_locker

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	zeroLogger "github.com/rs/zerolog/log"
)

// RedisLocker implements FlowLocker using Redis with Redlock
type RedisLocker struct {
	rs     *redsync.Redsync
	expiry time.Duration
}

// RedisLockerConfig holds configuration for RedisLocker
type RedisLockerConfig struct {
	Address  string
	Password string
	Expiry   time.Duration
}

// NewRedisLocker creates a new Redis-based flow locker
func NewRedisLocker(config RedisLockerConfig) *RedisLocker {
	if config.Expiry == 0 {
		config.Expiry = 30 * time.Second
	}

	// Create redigo pool
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Address)
			if err != nil {
				return nil, err
			}
			if config.Password != "" {
				if _, err := c.Do("AUTH", config.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
	}

	// Create redsync instance
	rs := redsync.New(redigo.NewPool(pool))

	return &RedisLocker{
		rs:     rs,
		expiry: config.Expiry,
	}
}

// Lock acquires a distributed lock for the given flow ID
func (r *RedisLocker) Lock(ctx context.Context, flowID uuid.UUID) (func(), error) {
	// Check if context is already canceled before attempting lock
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	mutex := r.rs.NewMutex(
		"flow:lock:"+flowID.String(),
		redsync.WithExpiry(r.expiry),
		redsync.WithTries(1),
	)

	if err := mutex.LockContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	zeroLogger.Debug().
		Str("flow_id", flowID.String()).
		Msg("acquired distributed flow lock")

	unlock := func() {
		if ok, err := mutex.UnlockContext(context.Background()); !ok || err != nil {
			zeroLogger.Error().
				Err(err).
				Str("flow_id", flowID.String()).
				Msg("failed to release lock")
		} else {
			zeroLogger.Debug().
				Str("flow_id", flowID.String()).
				Msg("released distributed flow lock")
		}
	}

	return unlock, nil
}
