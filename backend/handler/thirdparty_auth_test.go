package handler

import (
	"github.com/teamhanko/hanko/backend/thirdparty"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func (s *thirdPartySuite) TestThirdPartyHandler_Auth() {
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
			name:                "successful redirect to apple",
			referer:             "https://login.test.example",
			enabledProviders:    []string{"apple"},
			allowedRedirectURLs: []string{"https://*.test.example"},
			requestedProvider:   "apple",
			requestedRedirectTo: "https://app.test.example",
			expectedBaseURL:     "https://" + thirdparty.AppleAPIBase + thirdparty.AppleAuthEndpoint,
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
		s.Run(testData.name, func() {
			cfg := s.setUpConfig(testData.enabledProviders, testData.allowedRedirectURLs)

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

			c, rec := s.setUpContext(req)
			handler := s.setUpHandler(cfg)

			err := handler.Auth(c)
			s.NoError(err)

			s.Equal(http.StatusTemporaryRedirect, rec.Code)

			u, err := url.Parse(rec.Header().Get("Location"))
			s.NoError(err)

			s.Equal(testData.expectedBaseURL, u.Scheme+"://"+u.Host+u.Path)

			q := u.Query()

			if testData.expectedError != "" {
				s.Equal(testData.expectedError, q.Get("error"))
				errorDescription := q.Get("error_description")
				isCorrectErrorDescription := strings.Contains(errorDescription, testData.expectedErrorDescription)
				s.Truef(isCorrectErrorDescription, "error description '%s' does not contain '%s'", errorDescription, testData.expectedErrorDescription)
			} else {
				s.Equal(cfg.ThirdParty.RedirectURL, q.Get("redirect_uri"))
				s.Equal(cfg.ThirdParty.Providers.Get(testData.requestedProvider).ClientID, q.Get("client_id"))
				s.Equal("code", q.Get("response_type"))

				expectedState := rec.Result().Cookies()[0].Value
				state, err := thirdparty.VerifyState(cfg, q.Get("state"), expectedState)
				s.NoError(err)

				s.Equal(strings.ToLower(testData.requestedProvider), state.Provider)

				if testData.requestedRedirectTo == "" {
					s.Equal(cfg.ThirdParty.ErrorRedirectURL, state.RedirectTo)
				} else {
					s.Equal(testData.requestedRedirectTo, state.RedirectTo)
				}
			}
		})
	}
}
