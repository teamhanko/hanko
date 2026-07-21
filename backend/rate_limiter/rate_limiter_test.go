package rate_limiter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-limiter"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/utils"
)

func TestNewRateLimiter(t *testing.T) {
	cfg := config.RateLimiter{
		Enabled: true,
		Store:   config.RATE_LIMITER_STORE_IN_MEMORY,
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

func TestLimit_IPExtractorRateLimitBehavior(t *testing.T) {
	tests := []struct {
		name                   string
		ipConfig               config.IPConfig
		firstRemoteAddr        string
		firstHeaders           map[string]string
		secondRemoteAddr       string
		secondHeaders          map[string]string
		expectSecondRateLimit  bool
		expectedSecondErrorMsg string
	}{
		{
			name: "direct extractor cannot be bypassed with spoofed X-Forwarded-For",
			ipConfig: config.IPConfig{
				Extractor: config.IPExtractorDirect,
			},
			firstRemoteAddr: "198.51.100.10:11111",
			firstHeaders: map[string]string{
				echo.HeaderXForwardedFor: "203.0.113.1",
			},
			secondRemoteAddr: "198.51.100.10:11111",
			secondHeaders: map[string]string{
				echo.HeaderXForwardedFor: "203.0.113.2",
			},
			expectSecondRateLimit: true,
		},
		{
			name: "X-Forwarded-For extractor uses header from trusted proxy",
			ipConfig: config.IPConfig{
				Extractor:      config.IPExtractorXForwardedFor,
				TrustedProxies: []string{"10.0.0.0/8"},
			},
			firstRemoteAddr: "10.0.1.25:11111",
			firstHeaders: map[string]string{
				echo.HeaderXForwardedFor: "203.0.113.1",
			},
			secondRemoteAddr: "10.0.1.25:22222",
			secondHeaders: map[string]string{
				echo.HeaderXForwardedFor: "203.0.113.2",
			},
			expectSecondRateLimit: false,
		},
		{
			name: "X-Forwarded-For extractor ignores header from untrusted peer",
			ipConfig: config.IPConfig{
				Extractor:      config.IPExtractorXForwardedFor,
				TrustedProxies: []string{"10.0.0.0/8"},
			},
			firstRemoteAddr: "198.51.100.10:11111",
			firstHeaders: map[string]string{
				echo.HeaderXForwardedFor: "203.0.113.1",
			},
			secondRemoteAddr: "198.51.100.10:22222",
			secondHeaders: map[string]string{
				echo.HeaderXForwardedFor: "203.0.113.2",
			},
			expectSecondRateLimit: true,
		},
		{
			name: "X-Real-IP extractor uses header from trusted proxy",
			ipConfig: config.IPConfig{
				Extractor:      config.IPExtractorXRealIP,
				TrustedProxies: []string{"10.0.0.0/8"},
			},
			firstRemoteAddr: "10.0.1.25:11111",
			firstHeaders: map[string]string{
				echo.HeaderXRealIP: "203.0.113.1",
			},
			secondRemoteAddr: "10.0.1.25:22222",
			secondHeaders: map[string]string{
				echo.HeaderXRealIP: "203.0.113.2",
			},
			expectSecondRateLimit: false,
		},
		{
			name: "X-Real-IP extractor ignores header from untrusted peer",
			ipConfig: config.IPConfig{
				Extractor:      config.IPExtractorXRealIP,
				TrustedProxies: []string{"10.0.0.0/8"},
			},
			firstRemoteAddr: "198.51.100.10:11111",
			firstHeaders: map[string]string{
				echo.HeaderXRealIP: "203.0.113.1",
			},
			secondRemoteAddr: "198.51.100.10:22222",
			secondHeaders: map[string]string{
				echo.HeaderXRealIP: "203.0.113.2",
			},
			expectSecondRateLimit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			err := utils.ConfigureIPExtractor(e, tt.ipConfig)
			if err != nil {
				t.Fatal(err)
			}

			store := newTestRateLimiter(1)
			userID := uuid.Must(uuid.NewV4())

			firstContext := newRateLimitTestContext(e, http.MethodPost, "/login", tt.firstRemoteAddr, tt.firstHeaders)
			if err := Limit(store, userID, firstContext); err != nil {
				t.Fatalf("first request should not be rate limited: %v", err)
			}

			secondContext := newRateLimitTestContext(e, http.MethodPost, "/login", tt.secondRemoteAddr, tt.secondHeaders)
			secondErr := Limit(store, userID, secondContext)

			if tt.expectSecondRateLimit {
				assertTooManyRequests(t, secondErr)
				return
			}

			if secondErr != nil {
				t.Fatalf("second request should not be rate limited: %v", secondErr)
			}
		})
	}
}

func newTestRateLimiter(tokens uint64) limiter.Store {
	return NewRateLimiter(
		config.RateLimiter{
			Enabled: true,
			Store:   config.RATE_LIMITER_STORE_IN_MEMORY,
			Redis:   nil,
		},
		config.RateLimits{
			Tokens:   tokens,
			Interval: time.Minute,
		},
	)
}

func newRateLimitTestContext(e *echo.Echo, method, target, remoteAddr string, headers map[string]string) echo.Context {
	req := httptest.NewRequest(method, target, nil)
	req.RemoteAddr = remoteAddr

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(target)

	return c
}

func assertTooManyRequests(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected rate limit error")
	}

	httpError, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}

	if httpError.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, httpError.Code)
	}
}
