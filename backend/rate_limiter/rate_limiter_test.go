package rate_limiter

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/config"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	cfg := config.RateLimiter{
		Enabled: true,
		Backend: config.RATE_LIMITER_BACKEND_IN_MEMORY,
		Redis:   nil,
	}

	rl := NewRateLimiter(cfg, config.RateLimits{
		Tokens:   5,
		Interval: 1 * time.Minute,
	})
	// Take 5 tokens: should be good.
	for i := 0; i < 5; i++ {
		tokens, remaining, reset, ok, e := rl.Take(context.Background(), "some-key")
		log.Printf("Tokens: %v, Remaining: %v, Reset: %v, ok: %v, error: %v\n", tokens, remaining, time.Unix(0, int64(reset)).String(), ok, e)
		if e != nil {
			t.Error(e)
		}
		if !ok {
			t.Error("Taking a token should succeed at this point.")
		}
	}

	tokens, remaining, reset, ok, e := rl.Take(context.Background(), "some-key")
	log.Printf("Tokens: %v, Remaining: %v, Reset: %v, ok: %v, error: %v\n", tokens, remaining, time.Unix(0, int64(reset)).String(), ok, e)
	if ok {
		t.Error("Taking a token should fail at this point")
	}
}
