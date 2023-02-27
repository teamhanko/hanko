package thirdparty

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"net/url"
	"strings"
	"testing"
)

func TestIsValidRedirectTo(t *testing.T) {
	tests := []struct {
		name                string
		requestedRedirect   string
		allowedRedirectURLs []string
		errorRedirectURL    string
	}{
		{
			name:                "Exact match",
			requestedRedirect:   "https://foo.example.com",
			allowedRedirectURLs: []string{"https://foo.example.com"},
		},
		{
			name:                "Subdomain match",
			requestedRedirect:   "https://foo.example.com",
			allowedRedirectURLs: []string{"https://*.example.com"},
		},
		{
			name:                "Path match",
			requestedRedirect:   "https://foo.example.com/page/anotherPage",
			allowedRedirectURLs: []string{"https://foo.example.com/page/anotherPage"},
		},
		{
			name:                "Trailing slash ignored",
			requestedRedirect:   "https://foo.example.com/",
			allowedRedirectURLs: []string{"https://*.example.com"},
		},
		{
			name:              "Error redirect url, trailing slash ignored",
			requestedRedirect: "https://foo.example.com/error/",
			errorRedirectURL:  "https://foo.example.com/error",
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
			assert.True(t, got)
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
			expectedError:            ThirdPartyErrorCodeServerError,
			expectedErrorDescription: "an internal error has occurred",
		},
		{
			name:                     "return url with third party error code and description",
			redirectTo:               "https://foo.example.com",
			error:                    ErrorUserConflict("user already exists"),
			expectedError:            ThirdPartyErrorCodeUserConflict,
			expectedErrorDescription: "user already exists",
		},
		{
			name:                     "return url with server error when error is not a third party error",
			redirectTo:               "https://foo.example.com",
			error:                    errors.New("non-third party error"),
			expectedError:            ThirdPartyErrorCodeServerError,
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
