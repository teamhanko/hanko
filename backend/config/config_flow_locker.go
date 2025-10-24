package config

import (
	"errors"
	"time"
)

type FlowLocker struct {
	// `enabled` controls whether flow locking is enabled
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `store` sets the backend for the flow locker
	Store FlowLockerStoreType `yaml:"store" json:"store,omitempty" koanf:"store" jsonschema:"default=in_memory,enum=in_memory,enum=redis"`
	// `redis_config` configures connection to a redis instance
	// Required if `store` is set to `redis`
	Redis *RedisConfig `yaml:"redis_config" json:"redis_config,omitempty" koanf:"redis_config"`
	// `ttl` is the lock timeout (for Redis only)
	TTL time.Duration `yaml:"ttl" json:"ttl,omitempty" koanf:"ttl" jsonschema:"default=30s,type=string"`
}

type FlowLockerStoreType string

const (
	FLOW_LOCKER_STORE_IN_MEMORY FlowLockerStoreType = "in_memory"
	FLOW_LOCKER_STORE_REDIS     FlowLockerStoreType = "redis"
)

func (f *FlowLocker) Validate() error {
	if f.Enabled {
		switch f.Store {
		case FLOW_LOCKER_STORE_REDIS:
			if f.Redis == nil {
				return errors.New("when enabling the redis store you have to specify the redis config")
			}
			if f.Redis.Address == "" {
				return errors.New("when enabling the redis store you have to specify the address where hanko can reach the redis instance")
			}
		case FLOW_LOCKER_STORE_IN_MEMORY:
			break
		default:
			return errors.New(string(f.Store) + " is not a valid flow locker store")
		}
	}
	return nil
}
