package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestThirdPartyHandler_Auth(t *testing.T) {

	tests := []struct {
		name                     string
		referer                  string
		enabledProviders         []string
		allowedRedirectURLs      []string
		requestedProvider        string
		requestedRedirectTo      string
		expectedBaseURL          string
		expectedError            string
		expectedErrorDescription string // can be a partial message
	}{
		{
			name:                "successful redirect to google",
			referer:             "https://login.test.example",
			enabledProviders:    []string{"google"},
			allowedRedirectURLs: []string{"https://*.test.example"},
			requestedProvider:   "google",
			requestedRedirectTo: "https://app.test.example",
			expectedBaseURL:     "https://" + thirdparty.GoogleAuthBase + thirdparty.GoogleOauthAuthEndpoint,
		},
		{
			name:                "successful redirect to github",
			referer:             "https://login.test.example",
			enabledProviders:    []string{"github"},
			allowedRedirectURLs: []string{"https://*.test.example"},
			requestedProvider:   "github",
			requestedRedirectTo: "https://app.test.example",
			expectedBaseURL:     "https://" + thirdparty.GithubAuthBase + thirdparty.GithubOauthAuthEndpoint,
		},
		{
			name:                     "error redirect on missing provider",
			referer:                  "https://login.test.example",
			requestedRedirectTo:      "https://app.test.example",
			expectedBaseURL:          "https://login.test.example",
			expectedError:            thirdparty.ThirdPartyErrorCodeInvalidRequest,
			expectedErrorDescription: "is a required field",
		},
		{
			name:                     "error redirect on missing redirectTo",
			referer:                  "https://login.test.example",
			requestedProvider:        "google",
			expectedBaseURL:          "https://login.test.example",
			expectedError:            thirdparty.ThirdPartyErrorCodeInvalidRequest,
			expectedErrorDescription: "is a required field",
		},
		{
			name:                     "error redirect when requested provider is disabled",
			referer:                  "https://login.test.example",
			enabledProviders:         []string{"github"},
			allowedRedirectURLs:      []string{"https://*.test.example"},
			requestedProvider:        "google",
			requestedRedirectTo:      "https://app.test.example",
			expectedBaseURL:          "https://login.test.example",
			expectedError:            thirdparty.ThirdPartyErrorCodeInvalidRequest,
			expectedErrorDescription: "provider is disabled",
		},
		{
			name:                     "error redirect when requesting an unknown provider",
			referer:                  "https://login.test.example",
			allowedRedirectURLs:      []string{"https://*.test.example"},
			requestedProvider:        "unknownProvider",
			requestedRedirectTo:      "https://app.test.example",
			expectedBaseURL:          "https://login.test.example",
			expectedError:            thirdparty.ThirdPartyErrorCodeInvalidRequest,
			expectedErrorDescription: "is not supported",
		},
		{
			name:                     "error redirect when requesting a redirectTo that is not allowed",
			referer:                  "https://login.test.example",
			enabledProviders:         []string{"google"},
			allowedRedirectURLs:      []string{"https://*.test.example"},
			requestedProvider:        "google",
			requestedRedirectTo:      "https://app.test.wrong",
			expectedBaseURL:          "https://login.test.example",
			expectedError:            thirdparty.ThirdPartyErrorCodeInvalidRequest,
			expectedErrorDescription: "redirect to 'https://app.test.wrong' not allowed",
		},
		{
			name:                     "error redirect with redirect to error redirect url if referer not present",
			allowedRedirectURLs:      []string{"https://*.test.example"},
			requestedProvider:        "unknownProvider",
			requestedRedirectTo:      "https://app.test.example",
			expectedBaseURL:          "https://error.test.example",
			expectedError:            thirdparty.ThirdPartyErrorCodeInvalidRequest,
			expectedErrorDescription: "is not supported",
		},
	}

	for _, testData := range tests {
		t.Run(testData.name, func(t *testing.T) {
			cfg := setUpConfig(t, testData.enabledProviders, testData.allowedRedirectURLs)
			e := echo.New()
			e.Validator = dto.NewCustomValidator()

			req := httptest.NewRequest(http.MethodGet, "/thirdparty/auth", nil)

			params := url.Values{}
			if testData.requestedProvider != "" {
				params.Add("provider", testData.requestedProvider)
			}
			if testData.requestedRedirectTo != "" {
				params.Add("redirect_to", testData.requestedRedirectTo)
			}
			req.URL.RawQuery = params.Encode()

			req.Header.Set("Referer", testData.referer)

			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

			jwkManager := test.JwkManager{}
			sessionMgr, err := session.NewManager(jwkManager, cfg.Session)
			require.NoError(t, err)

			handler := NewThirdPartyHandler(cfg, p, sessionMgr, test.NewAuditLogger())

			err = handler.Auth(c)
			require.NoError(t, err)

			assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)

			u, err := url.Parse(rec.Header().Get("Location"))
			assert.NoError(t, err, "redirect url parse failed")

			assert.Equal(t, testData.expectedBaseURL, u.Scheme+"://"+u.Host+u.Path)

			q := u.Query()

			if testData.expectedError != "" {
				assert.Equal(t, testData.expectedError, q.Get("error"))
				errorDescription := q.Get("error_description")
				isCorrectErrorDescription := strings.Contains(errorDescription, testData.expectedErrorDescription)
				assert.Truef(t, isCorrectErrorDescription, "error description '%s' does not contain '%s'", errorDescription, testData.expectedErrorDescription)
			} else {
				assert.Equal(t, cfg.ThirdParty.RedirectURL, q.Get("redirect_uri"))
				assert.Equal(t, cfg.ThirdParty.Providers.Get(testData.requestedProvider).ClientID, q.Get("client_id"))
				assert.Equal(t, "code", q.Get("response_type"))

				expectedState := rec.Result().Cookies()[0].Value
				state, err := thirdparty.VerifyState(cfg, q.Get("state"), expectedState)
				require.NoError(t, err)

				assert.Equal(t, strings.ToLower(testData.requestedProvider), state.Provider)

				if testData.requestedRedirectTo == "" {
					assert.Equal(t, cfg.ThirdParty.ErrorRedirectURL, state.RedirectTo)
				} else {
					assert.Equal(t, testData.requestedRedirectTo, state.RedirectTo)
				}
			}
		})
	}
}

func setUpConfig(t *testing.T, enabledProviders []string, allowedRedirectURLs []string) *config.Config {
	cfg := &config.Config{ThirdParty: config.ThirdParty{
		Providers: config.ThirdPartyProviders{
			Google: config.ThirdPartyProvider{
				Enabled:  false,
				ClientID: "fakeClientID",
				Secret:   "fakeClientSecret",
			}, GitHub: config.ThirdPartyProvider{
				Enabled:  false,
				ClientID: "fakeClientID",
				Secret:   "fakeClientSecret",
			}},
		ErrorRedirectURL:    "https://error.test.example",
		RedirectURL:         "https://api.test.example/callback",
		AllowedRedirectURLS: allowedRedirectURLs,
	}, Secrets: config.Secrets{Keys: []string{"thirty-two-byte-long-test-secret"}}}

	for _, provider := range enabledProviders {
		switch provider {
		case "google":
			cfg.ThirdParty.Providers.Google.Enabled = true
		case "github":
			cfg.ThirdParty.Providers.GitHub.Enabled = true
		}
	}

	err := cfg.PostProcess()
	require.NoError(t, err)

	return cfg
}
