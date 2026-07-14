package thirdparty

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v3/config"
)

func TestIsValidRedirectTo(t *testing.T) {
	tests := []struct {
		name                string
		requestedRedirect   string
		allowedRedirectURLs []string
		errorRedirectURL    string
		want                bool
	}{
		// --- existing positive cases ---
		{
			name:                "Exact match",
			requestedRedirect:   "https://foo.example.com",
			allowedRedirectURLs: []string{"https://foo.example.com"},
			want:                true,
		},
		{
			name:                "Subdomain match",
			requestedRedirect:   "https://foo.example.com",
			allowedRedirectURLs: []string{"https://*.example.com"},
			want:                true,
		},
		{
			name:                "Path match",
			requestedRedirect:   "https://foo.example.com/page/anotherPage",
			allowedRedirectURLs: []string{"https://foo.example.com/page/anotherPage"},
			want:                true,
		},
		{
			name:                "Trailing slash ignored",
			requestedRedirect:   "https://foo.example.com/",
			allowedRedirectURLs: []string{"https://*.example.com"},
			want:                true,
		},
		{
			name:                "Starts with super glob",
			requestedRedirect:   "https://foo.example.com",
			allowedRedirectURLs: []string{"https://**.example.com"},
			want:                true,
		},
		{
			name:              "Error redirect url, trailing slash ignored",
			requestedRedirect: "https://foo.example.com/error/",
			errorRedirectURL:  "https://foo.example.com/error",
			want:              true,
		},

		// --- empty input ---
		{
			name:                "Empty redirectTo rejected",
			requestedRedirect:   "",
			allowedRedirectURLs: []string{"https://example.com"},
			want:                false,
		},

		// --- relative and protocol-relative URLs ---
		{
			name:                "Relative URL rejected",
			requestedRedirect:   "/foo",
			allowedRedirectURLs: []string{"https://example.com"},
			want:                false,
		},
		{
			name:                "Protocol-relative URL rejected",
			requestedRedirect:   "//evil.com",
			allowedRedirectURLs: []string{"https://evil.com"},
			want:                false,
		},

		// --- missing hostname ---
		{
			name:                "URL with empty hostname rejected",
			requestedRedirect:   "https:///broken",
			allowedRedirectURLs: []string{"https:///broken"},
			want:                false,
		},

		// --- forbidden schemes ---
		{
			name:                "javascript: scheme rejected",
			requestedRedirect:   "javascript:alert(1)",
			allowedRedirectURLs: []string{"javascript:alert(1)"},
			want:                false,
		},
		{
			name:                "data: scheme rejected",
			requestedRedirect:   "data:text/html,hi",
			allowedRedirectURLs: []string{"data:text/html,**"},
			want:                false,
		},
		{
			name:                "file: scheme rejected",
			requestedRedirect:   "file:///etc/passwd",
			allowedRedirectURLs: []string{"file:///etc/passwd"},
			want:                false,
		},

		// --- host-collision prevention ---
		{
			name:                "IP glob does not match attacker host",
			requestedRedirect:   "http://127.0.0.1.evil.com",
			allowedRedirectURLs: []string{"http://127.0.0.1**"},
			want:                false,
		},
		{
			name:                "IP glob allows exact IP",
			requestedRedirect:   "http://127.0.0.1",
			allowedRedirectURLs: []string{"http://127.0.0.1**"},
			want:                true,
		},
		{
			name:                "IP glob allows path on exact IP",
			requestedRedirect:   "http://127.0.0.1/dashboard",
			allowedRedirectURLs: []string{"http://127.0.0.1**"},
			want:                true,
		},
		{
			name:                "localhost glob does not allow localhost.evil.com",
			requestedRedirect:   "http://localhost.evil.com",
			allowedRedirectURLs: []string{"http://localhost**"},
			want:                false,
		},
		{
			name:                "Super glob has path suffix",
			requestedRedirect:   "https://example.com.evil.com/test",
			allowedRedirectURLs: []string{"https://example.com**/test"},
			want:                false,
		},

		// --- mid-host wildcard subdomain ---
		{
			name:                "Mid-host wildcard subdomain match",
			requestedRedirect:   "https://foo.mid.bar.com",
			allowedRedirectURLs: []string{"https://foo.*.bar.com"},
			want:                true,
		},
		{
			name:                "Mid-host wildcard subdomain rejects extra nested level",
			requestedRedirect:   "https://foo.a.b.bar.com",
			allowedRedirectURLs: []string{"https://foo.*.bar.com"},
			want:                false,
		},

		// --- dash-suffixed label wildcard  ---
		{
			name:                "Dash-suffixed label wildcard match",
			requestedRedirect:   "https://foo-prod.bar.com",
			allowedRedirectURLs: []string{"https://foo-*.bar.com"},
			want:                true,
		},
		{
			name:                "Dash-suffixed label wildcard rejects wrong suffix domain",
			requestedRedirect:   "https://foo-prod.evil.com",
			allowedRedirectURLs: []string{"https://foo-*.bar.com"},
			want:                false,
		},

		// --- IP subnet wildcard  ---
		{
			name:                "IP subnet wildcard matches address in range",
			requestedRedirect:   "http://192.168.1.55/dashboard",
			allowedRedirectURLs: []string{"http://192.168.*.*/**"},
			want:                true,
		},
		{
			name:                "IP subnet wildcard rejects address outside range",
			requestedRedirect:   "http://10.0.0.1/dashboard",
			allowedRedirectURLs: []string{"http://192.168.*.*/**"},
			want:                false,
		},
		{
			name:                "IP subnet wildcard rejects host-collision with extra label",
			requestedRedirect:   "http://192.168.1.55.evil.com/dashboard",
			allowedRedirectURLs: []string{"http://192.168.*.*/**"},
			want:                false,
		},

		// --- anchored mid-pattern super-glob ---
		{
			name:                "Anchored mid-pattern super-glob allows nested subdomains",
			requestedRedirect:   "https://foo.a.b.bar.com",
			allowedRedirectURLs: []string{"https://foo.**.bar.com"},
			want:                true,
		},
		{
			name:                "Anchored mid-pattern super-glob still requires literal suffix domain",
			requestedRedirect:   "https://foo.a.b.evil.com",
			allowedRedirectURLs: []string{"https://foo.**.bar.com"},
			want:                false,
		},

		// --- super-wildcard subdomain depth ---
		{
			name:                "Super-wildcard subdomain allows nested multi-level subdomains",
			requestedRedirect:   "https://a.b.example.com",
			allowedRedirectURLs: []string{"https://**.example.com"},
			want:                true,
		},
		{
			name:                "Super-wildcard subdomain does not match bare base domain",
			requestedRedirect:   "https://example.com",
			allowedRedirectURLs: []string{"https://**.example.com"},
			want:                false,
		},

		// --- subdomain wildcard combined with a bare trailing ** ---
		{
			name:                "Subdomain wildcard with trailing super glob matches subdomain",
			requestedRedirect:   "https://app.gett.co/callback",
			allowedRedirectURLs: []string{"https://*.gett.co**"},
			want:                true,
		},
		{
			name:                "Subdomain wildcard with trailing super glob rejects base domain",
			requestedRedirect:   "https://gett.co",
			allowedRedirectURLs: []string{"https://*.gett.co**"},
			want:                false,
		},
		{
			name:                "Super-wildcard subdomain with trailing super glob matches nested subdomain",
			requestedRedirect:   "https://a.b.kyr.link/callback",
			allowedRedirectURLs: []string{"https://**.kyr.link**"},
			want:                true,
		},
		{
			name:                "Subdomain wildcard with trailing super glob rejects host-collision",
			requestedRedirect:   "https://app.gett.co.evil.com",
			allowedRedirectURLs: []string{"https://*.gett.co**"},
			want:                false,
		},

		// --- regression guards: a trailing wildcard after ** is not a safe anchor ---
		{
			name:                "Trailing bracket class after ** does not create a safe anchor",
			requestedRedirect:   "http://127.0.0.1.evil.com",
			allowedRedirectURLs: []string{"http://127.0.0.1**[a-z]"},
			want:                false,
		},
		{
			name:                "Trailing dot-star after ** does not create a safe anchor",
			requestedRedirect:   "http://127.0.0.1.evil.com",
			allowedRedirectURLs: []string{"http://127.0.0.1**.*"},
			want:                false,
		},

		// --- subdomain wildcard boundary ---
		{
			name:                "Wildcard matches real subdomain",
			requestedRedirect:   "https://app.example.com",
			allowedRedirectURLs: []string{"https://*.example.com"},
			want:                true,
		},
		{
			name:                "Wildcard does not match base domain",
			requestedRedirect:   "https://example.com",
			allowedRedirectURLs: []string{"https://*.example.com"},
			want:                false,
		},
		{
			name:                "Wildcard does not match sibling with base as suffix",
			requestedRedirect:   "https://badexample.com",
			allowedRedirectURLs: []string{"https://*.example.com"},
			want:                false,
		},
		{
			name:                "Wildcard does not match evil suffix on base domain",
			requestedRedirect:   "https://example.com.evil.com",
			allowedRedirectURLs: []string{"https://*.example.com"},
			want:                false,
		},
		{
			name:                "Wildcard does not match nested subdomain (glob separator blocks it)",
			requestedRedirect:   "https://a.b.example.com",
			allowedRedirectURLs: []string{"https://*.example.com"},
			want:                false,
		},

		// --- port matching ---
		{
			name:                "Correct port allowed",
			requestedRedirect:   "http://localhost:8888/foo",
			allowedRedirectURLs: []string{"http://localhost:8888**"},
			want:                true,
		},
		{
			name:                "Wrong port rejected",
			requestedRedirect:   "http://localhost:9999/foo",
			allowedRedirectURLs: []string{"http://localhost:8888**"},
			want:                false,
		},
		{
			name:                "Port omitted in pattern — any port accepted",
			requestedRedirect:   "http://localhost/foo",
			allowedRedirectURLs: []string{"http://localhost**"},
			want:                true,
		},
		{
			name:                "Port in pattern required — URL without port rejected",
			requestedRedirect:   "http://localhost/foo",
			allowedRedirectURLs: []string{"http://localhost:8888**"},
			want:                false,
		},

		// --- scheme mismatch ---
		{
			name:                "http pattern does not allow https redirect",
			requestedRedirect:   "https://example.com",
			allowedRedirectURLs: []string{"http://example.com"},
			want:                false,
		},
		{
			name:                "https pattern does not allow http redirect",
			requestedRedirect:   "http://example.com",
			allowedRedirectURLs: []string{"https://example.com"},
			want:                false,
		},

		// --- path glob ---
		{
			name:                "Double-star glob matches any path",
			requestedRedirect:   "https://example.com/deep/path/page",
			allowedRedirectURLs: []string{"https://example.com/**"},
			want:                true,
		},
		{
			name:                "Single-star glob matches one path segment",
			requestedRedirect:   "https://example.com/page",
			allowedRedirectURLs: []string{"https://example.com/*"},
			want:                true,
		},
		{
			name:                "Single-star glob does not match multi-segment path",
			requestedRedirect:   "https://example.com/page/sub",
			allowedRedirectURLs: []string{"https://example.com/*"},
			want:                false,
		},

		// --- no matching entry ---
		{
			name:                "No allowlist entry matches",
			requestedRedirect:   "https://other.com",
			allowedRedirectURLs: []string{"https://example.com"},
			want:                false,
		},
		{
			name:                "Empty allowlist",
			requestedRedirect:   "https://example.com",
			allowedRedirectURLs: []string{},
			want:                false,
		},
	}

	for _, testData := range tests {
		t.Run(testData.name, func(t *testing.T) {
			cfg := config.ThirdParty{
				AllowedRedirectURLS: testData.allowedRedirectURLs,
			}

			if testData.errorRedirectURL != "" {
				cfg.ErrorRedirectURL = testData.errorRedirectURL
			}

			err := cfg.PostProcess()
			require.NoError(t, err)

			got := IsAllowedRedirect(cfg, testData.requestedRedirect)
			assert.Equal(t, testData.want, got)
		})
	}
}

func TestGetErrorUrl(t *testing.T) {
	tests := []struct {
		name                     string
		redirectTo               string
		error                    error
		expectedError            string
		expectedErrorDescription string
	}{
		{
			name:                     "return url with server error when error is a third party server or invalid request error",
			redirectTo:               "https://foo.example.com",
			error:                    ErrorServer("could not decode payload"),
			expectedError:            ErrorCodeServerError,
			expectedErrorDescription: "an internal error has occurred",
		},
		{
			name:                     "return url with third party error code and description",
			redirectTo:               "https://foo.example.com",
			error:                    ErrorUserConflict("user already exists"),
			expectedError:            ErrorCodeUserConflict,
			expectedErrorDescription: "user already exists",
		},
		{
			name:                     "return url with server error when error is not a third party error",
			redirectTo:               "https://foo.example.com",
			error:                    errors.New("non-third party error"),
			expectedError:            ErrorCodeServerError,
			expectedErrorDescription: "an internal error has occurred",
		},
	}

	for _, testData := range tests {
		t.Run(testData.name, func(t *testing.T) {
			got := GetErrorUrl(testData.redirectTo, testData.error)

			u, err := url.Parse(got)
			require.NoError(t, err)

			assert.Equal(t, testData.redirectTo, u.Scheme+"://"+u.Host)
			assert.Equal(t, testData.expectedError, u.Query().Get("error"))
			errorDescription := u.Query().Get("error_description")
			isCorrectErrorDescription := strings.Contains(errorDescription, testData.expectedErrorDescription)
			assert.Truef(t, isCorrectErrorDescription, "error description '%s' does not contain '%s'", errorDescription, testData.expectedErrorDescription)
		})
	}
}
