package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v3/config"
)

func TestConfigureSchemeExtractor_SchemeResolution(t *testing.T) {
	tests := []struct {
		name           string
		cfg            config.IPConfig
		remoteAddr     string
		expectedScheme string
	}{
		{
			name:           "private peer is trusted by default",
			cfg:            config.IPConfig{},
			remoteAddr:     "10.0.1.25:12345",
			expectedScheme: "https",
		},
		{
			name:           "public peer is ignored by default",
			cfg:            config.IPConfig{},
			remoteAddr:     "35.191.1.2:12345",
			expectedScheme: "http",
		},
		{
			name:           "public peer in trusted_proxies is trusted",
			cfg:            config.IPConfig{TrustedProxies: []string{"35.191.0.0/16"}},
			remoteAddr:     "35.191.1.2:12345",
			expectedScheme: "https",
		},
		{
			name:           "public peer outside trusted_proxies is ignored",
			cfg:            config.IPConfig{TrustedProxies: []string{"35.191.0.0/16"}},
			remoteAddr:     "198.51.100.10:12345",
			expectedScheme: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			err := ConfigureSchemeExtractor(e, tt.cfg)
			require.NoError(t, err)

			e.GET("/scheme", func(c echo.Context) error {
				return c.String(http.StatusOK, c.Scheme())
			})

			req := httptest.NewRequest(http.MethodGet, "/scheme", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Forwarded-Proto", "https")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, tt.expectedScheme, rec.Body.String())
		})
	}
}

func TestConfigureIPExtractor_RealIPResolution(t *testing.T) {
	tests := []struct {
		name           string
		cfg            config.IPConfig
		remoteAddr     string
		headers        map[string]string
		expectedRealIP string
	}{
		{
			name: "direct ignores X-Forwarded-For and X-Real-IP",
			cfg: config.IPConfig{
				Extractor: config.IPExtractorDirect,
			},
			remoteAddr: "198.51.100.10:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.99",
				"X-Real-IP":       "203.0.113.88",
			},
			expectedRealIP: "198.51.100.10",
		},
		{
			name: "x_forwarded_for trusts header from trusted proxy",
			cfg: config.IPConfig{
				Extractor: config.IPExtractorXForwardedFor,
				TrustedProxies: []string{
					"10.0.0.0/8",
				},
			},
			remoteAddr: "10.0.1.25:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.44",
			},
			expectedRealIP: "203.0.113.44",
		},
		{
			name: "x_forwarded_for ignores header from untrusted peer",
			cfg: config.IPConfig{
				Extractor: config.IPExtractorXForwardedFor,
				TrustedProxies: []string{
					"10.0.0.0/8",
				},
			},
			remoteAddr: "198.51.100.10:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.44",
			},
			expectedRealIP: "198.51.100.10",
		},
		{
			name: "x_real_ip trusts header from trusted proxy",
			cfg: config.IPConfig{
				Extractor: config.IPExtractorXRealIP,
				TrustedProxies: []string{
					"10.0.0.0/8",
				},
			},
			remoteAddr: "10.0.1.25:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.55",
			},
			expectedRealIP: "203.0.113.55",
		},
		{
			name: "x_real_ip ignores header from untrusted peer",
			cfg: config.IPConfig{
				Extractor: config.IPExtractorXRealIP,
				TrustedProxies: []string{
					"10.0.0.0/8",
				},
			},
			remoteAddr: "198.51.100.10:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.55",
			},
			expectedRealIP: "198.51.100.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			err := ConfigureIPExtractor(e, tt.cfg)
			require.NoError(t, err)

			e.GET("/ip", func(c echo.Context) error {
				return c.String(http.StatusOK, c.RealIP())
			})

			req := httptest.NewRequest(http.MethodGet, "/ip", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, tt.expectedRealIP, rec.Body.String())
		})
	}
}
