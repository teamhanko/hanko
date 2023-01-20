package rate_limiter

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/config"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	aMinute := 1 * time.Minute
	var five uint64 = 5
	cfg := config.RateLimiter{
		Enabled:  true,
		Backend:  config.RATE_LIMITER_BACKEND_IN_MEMORY,
		Redis:    nil,
		Tokens:   &five,
		Interval: &aMinute,
	}

	rl := NewRateLimiter(cfg)
	// Take 5 tokens: should be good.
	for i := 0; i < 5; i++ {
		tokens, remaining, reset, ok, e := rl.Take(context.Background(), "some-key")
		log.Printf("Tokens: %v, Remaining: %v, Reset: %v, ok: %v, error: %v\n", tokens, remaining, time.Unix(0, int64(reset)).String(), ok, e)
		if e != nil {
			t.Error(e)
		}
	}

	tokens, remaining, reset, ok, e := rl.Take(context.Background(), "some-key")
	log.Printf("Tokens: %v, Remaining: %v, Reset: %v, ok: %v, error: %v\n", tokens, remaining, time.Unix(0, int64(reset)).String(), ok, e)
	if ok {
		t.Error("Taking a token should fail at this point")
	}
}

func TestNewRateLimiterRedis(t *testing.T) {
	aMinute := 1 * time.Minute
	var five uint64 = 5
	cfg := config.RateLimiter{
		Enabled: true,
		Backend: config.RATE_LIMITER_BACKEND_REDIS,
		Redis: &config.RedisConfig{
			Address:  "localhost:6379",
			Password: "",
		},
		Tokens:   &five,
		Interval: &aMinute,
	}

	rl := NewRateLimiter(cfg)
	// Take 5 tokens: should be good.
	for i := 0; i < 6; i++ {
		tokens, remaining, e := rl.Get(context.Background(), "some-key")
		log.Printf("Tokens: %v, Remaining: %v\n", tokens, remaining)
		if e != nil {
			t.Error(e)
		}
	}

	// Try to take the sixth token, should faild
	_, _, e := rl.Get(context.Background(), "some-key")
	if e == nil {
		t.Error("Taking a token should fail at this point")
	}
}
