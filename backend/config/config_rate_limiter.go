package config

import (
	"errors"
	"time"
)

type RateLimiter struct {
	// `enabled` controls whether rate limiting is enabled or disabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `store` sets the store for the rate limiter. When you have multiple instances of Hanko running, it is recommended to use
	//  the `redis` store because otherwise your instances each have their own states.
	Store RateLimiterStoreType `yaml:"store" json:"store,omitempty" koanf:"store" jsonschema:"default=in_memory,enum=in_memory,enum=redis"`
	// `redis_config` configures connection to a redis instance.
	// Required if `store` is set to `redis`
	Redis *RedisConfig `yaml:"redis_config" json:"redis_config,omitempty" koanf:"redis_config"`
	// `passcode_limits` controls rate limits for passcode operations.
	PasscodeLimits RateLimits `yaml:"passcode_limits" json:"passcode_limits,omitempty" koanf:"passcode_limits" split_words:"true"`
	// `otp_limits` controls rate limits for OTP login attempts.
	OTPLimits RateLimits `yaml:"otp_limits" json:"otp_limits,omitempty" koanf:"otp_limits" split_words:"true"`
	// `password_limits` controls rate limits for password login operations.
	PasswordLimits RateLimits `yaml:"password_limits" json:"password_limits,omitempty" koanf:"password_limits" split_words:"true"`
	// `token_limits` controls rate limits for token exchange operations.
	TokenLimits RateLimits `yaml:"token_limits" json:"token_limits,omitempty" koanf:"token_limits" split_words:"true" jsonschema:"default=token=3;interval=1m"`
}

type RateLimits struct {
	// `tokens` determines how many operations/requests can occur in the given `interval`.
	Tokens uint64 `yaml:"tokens" json:"tokens" koanf:"tokens" jsonschema:"default=3"`
	// `interval` determines when to reset the token interval.
	// It must be a (possibly signed) sequence of decimal
	// numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Interval time.Duration `yaml:"interval" json:"interval" koanf:"interval" jsonschema:"default=1m,type=string"`
}

type RateLimiterStoreType string

const (
	RATE_LIMITER_STORE_IN_MEMORY RateLimiterStoreType = "in_memory"
	RATE_LIMITER_STORE_REDIS                          = "redis"
)

func (r *RateLimiter) Validate() error {
	if r.Enabled {
		switch r.Store {
		case RATE_LIMITER_STORE_REDIS:
			if r.Redis == nil {
				return errors.New("when enabling the redis store you have to specify the redis config")
			}
			if r.Redis.Address == "" {
				return errors.New("when enabling the redis store you have to specify the address where hanko can reach the redis instance")
			}
		case RATE_LIMITER_STORE_IN_MEMORY:
			break
		default:
			return errors.New(string(r.Store) + " is not a valid rate limiter store.")
		}
	}
	return nil
}

type RedisConfig struct {
	// `address` is the address of the redis instance in the form of `host[:port][/database]`.
	Address string `yaml:"address" json:"address" koanf:"address"`
	// `password` is the password for the redis instance.
	Password string `yaml:"password" json:"password,omitempty" koanf:"password"`
}
