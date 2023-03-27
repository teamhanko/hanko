package handler

import (
	"fmt"
	"github.com/h2non/gock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestThirdPartySuite(t *testing.T) {
	suite.Run(t, new(thirdPartySuite))
}

type thirdPartySuite struct {
	suite.Suite
	storage persistence.Storage
	db      *test.TestDB
}

func (s *thirdPartySuite) SetupSuite() {
	if testing.Short() {
		return
	}
	dialect := "postgres"
	db, err := test.StartDB("thirdparty_test", dialect)
	s.NoError(err)
	storage, err := persistence.New(config.Database{
		Url: db.DatabaseUrl,
	})
	s.NoError(err)

	s.storage = storage
	s.db = db
}

func (s *thirdPartySuite) SetupTest() {
	if s.db != nil {
		err := s.storage.MigrateUp()
		s.NoError(err)
	}
}

func (s *thirdPartySuite) TearDownTest() {
	if s.db != nil {
		err := s.storage.MigrateDown(-1)
		s.NoError(err)
	}
}

func (s *thirdPartySuite) TearDownSuite() {
	if s.db != nil {
		s.NoError(test.PurgeDB(s.db))
	}
}

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

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignUp_Google() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "test-google-signup@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("test-google-signup@example.com")
		s.NoError(err)
		s.NotNil(email)
		s.True(email.IsPrimary())

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("google", identity.ProviderName)
		s.Equal("google_abcde", identity.ProviderID)

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signup_succeeded"}, user.ID.String(), email.Address, "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignIn_Google() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "test-with-google-identity@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("test-with-google-identity@example.com")
		s.NoError(err)
		s.NotNil(email)
		s.True(email.IsPrimary())

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("google", identity.ProviderName)
		s.Equal("google_abcde", identity.ProviderID)

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_succeeded"}, user.ID.String(), "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignUp_GitHub() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	gock.New("https://" + thirdparty.GithubAuthBase).
		Post(thirdparty.GithubOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GithubAuthBase).
		Get(thirdparty.GithubUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GithubUser{
			ID:   1234,
			Name: "John Doe",
		})

	gock.New("https://" + thirdparty.GithubAPIBase).
		Get(thirdparty.GitHubEmailsEndpoint).
		Reply(200).
		JSON([]*thirdparty.GithubUserEmail{
			{
				Email:    "test-github-signup@example.com",
				Primary:  true,
				Verified: true,
			},
		})

	cfg := s.setUpConfig([]string{"github"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "github", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("test-github-signup@example.com")
		s.NoError(err)
		s.NotNil(email)
		s.True(email.IsPrimary())

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("github", identity.ProviderName)
		s.Equal("1234", identity.ProviderID)

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signup_succeeded"}, user.ID.String(), email.Address, "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignIn_GitHub() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GithubAuthBase).
		Post(thirdparty.GithubOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GithubAuthBase).
		Get(thirdparty.GithubUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GithubUser{
			ID:   1234,
			Name: "John Doe",
		})

	gock.New("https://" + thirdparty.GithubAPIBase).
		Get(thirdparty.GitHubEmailsEndpoint).
		Reply(200).
		JSON([]*thirdparty.GithubUserEmail{
			{
				Email:    "test-with-github-identity@example.com",
				Primary:  true,
				Verified: true,
			},
		})

	cfg := s.setUpConfig([]string{"github"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "github", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("test-with-github-identity@example.com")
		s.NoError(err)
		s.NotNil(email)
		s.True(email.IsPrimary())

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("github", identity.ProviderName)
		s.Equal("1234", identity.ProviderID)

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_succeeded"}, user.ID.String(), email.Address, "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignUp_WithUnclaimedEmail() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_unclaimed_email",
			Email:         "unclaimed-email@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("unclaimed-email@example.com")
		s.NoError(err)
		s.NotNil(email)
		s.True(email.IsPrimary())

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("google", identity.ProviderName)
		s.Equal("google_unclaimed_email", identity.ProviderID)

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signup_succeeded"}, user.ID.String(), email.Address, "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignIn_ProviderEMailChangedToExistingEmail() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "test-with-google-identity-changed@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("test-with-google-identity-changed@example.com")
		s.NoError(err)
		s.NotNil(email)

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("google", identity.ProviderName)
		s.Equal("google_abcde", identity.ProviderID)
		s.Equal("test-with-google-identity-changed@example.com", identity.Data["email"])

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_succeeded"}, user.ID.String(), user.Emails.GetPrimary().Address, "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignIn_ProviderEMailChangedToUnclaimedEmail() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "unclaimed-email@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("unclaimed-email@example.com")
		s.NoError(err)
		s.NotNil(email)

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("google", identity.ProviderName)
		s.Equal("google_abcde", identity.ProviderID)
		s.Equal("unclaimed-email@example.com", identity.Data["email"])

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_succeeded"}, user.ID.String(), user.Emails.GetPrimary().Address, "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_SignIn_ProviderEMailChangedToNonExistentEmail() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "non-existent-email@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)

		s.assertLocationHeaderHasToken(rec)
		s.assertStateCookieRemoved(rec)

		email, err := s.storage.GetEmailPersister().FindByAddress("non-existent-email@example.com")
		s.NoError(err)
		s.NotNil(email)

		user, err := s.storage.GetUserPersister().Get(*email.UserID)
		s.NoError(err)
		s.NotNil(user)
		s.Len(user.Emails, 3)

		identity := email.Identity
		s.NotNil(identity)
		s.Equal("google", identity.ProviderName)
		s.Equal("google_abcde", identity.ProviderID)
		s.Equal("non-existent-email@example.com", identity.Data["email"])

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_succeeded"}, user.ID.String(), user.Emails.GetPrimary().Address, "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_SignUpUserConflict() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_email_already_exists",
			Email:         "test-no-identity@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeUserConflict, location.Query().Get("error"))
		s.Equal("third party account linking for existing user with same email disallowed", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_SignInMultipleAccounts() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "provider-primary-email-changed@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeMultipleAccounts, location.Query().Get("error"))
		s.Equal(fmt.Sprintf("cannot identify associated user: '%s' is used by multiple accounts", "provider-primary-email-changed@example.com"), location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_NoState() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	req := httptest.NewRequest(http.MethodGet, "/thirdparty/callback?code=abcde", nil)

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("State is a required field", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_StateMismatch() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	mismatchedState, err := thirdparty.GenerateState(cfg, "github", "https://foo.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(mismatchedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("could not verify state", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_NoThirdPartyCookie() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("thirdparty state cookie is missing", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_ProviderError() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	providerError := "access_denied"
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s&error=%s", state, providerError), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(providerError, location.Query().Get("error"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_ProviderDisabled() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"github"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("google provider is disabled", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_NoAuthCode() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("auth code missing from request", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_OAuthTokenExchange() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(400).
		JSON(map[string]string{"error": "incorrect_client_credentials"})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("could not exchange authorization code for access token", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_VerificationRequiredUnverifiedProviderEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	gock.New("https://" + thirdparty.GoogleAuthBase).
		Post(thirdparty.GoogleOauthTokenEndpoint).
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New("https://" + thirdparty.GoogleAPIBase).
		Get(thirdparty.GoogleUserInfoEndpoint).
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "test-google-signup@example.com",
			EmailVerified: false,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})
	cfg.Emails.RequireVerification = true

	state, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", state), nil)
	req.AddCookie(&http.Cookie{
		Name:  HankoThirdpartyStateCookie,
		Value: string(state),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ThirdPartyErrorUnverifiedProviderEmail, location.Query().Get("error"))
		s.Equal("third party provider email must be verified", location.Query().Get("error_description"))

		logs, lerr := s.storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) setUpContext(request *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	rec := httptest.NewRecorder()
	c := e.NewContext(request, rec)
	return c, rec
}

func (s *thirdPartySuite) setUpHandler(cfg *config.Config) *ThirdPartyHandler {
	auditLogger := auditlog.NewLogger(s.storage, cfg.AuditLog)

	jwkMngr, err := jwk.NewDefaultManager(cfg.Secrets.Keys, s.storage.GetJwkPersister())
	s.Require().NoError(err)

	sessionMngr, err := session.NewManager(jwkMngr, cfg.Session)
	s.Require().NoError(err)

	handler := NewThirdPartyHandler(cfg, s.storage, sessionMngr, auditLogger)
	return handler
}

func (s *thirdPartySuite) setUpConfig(enabledProviders []string, allowedRedirectURLs []string) *config.Config {
	cfg := &config.Config{
		ThirdParty: config.ThirdParty{
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
		},
		Secrets: config.Secrets{
			Keys: []string{"thirty-two-byte-long-test-secret"},
		},
		AuditLog: config.AuditLog{
			Storage: config.AuditLogStorage{Enabled: true},
		},
		Emails: config.Emails{
			MaxNumOfAddresses: 5,
		},
	}

	for _, provider := range enabledProviders {
		switch provider {
		case "google":
			cfg.ThirdParty.Providers.Google.Enabled = true
		case "github":
			cfg.ThirdParty.Providers.GitHub.Enabled = true
		}
	}

	err := cfg.PostProcess()
	s.Require().NoError(err)

	return cfg
}

func (s *thirdPartySuite) assertLocationHeaderHasToken(rec *httptest.ResponseRecorder) {
	location, err := url.Parse(rec.Header().Get("Location"))
	s.NoError(err)
	s.True(location.Query().Has(HankoTokenQuery))
	s.NotEmpty(location.Query().Get(HankoTokenQuery))
}

func (s *thirdPartySuite) assertStateCookieRemoved(rec *httptest.ResponseRecorder) {
	cookies := rec.Result().Cookies()
	s.Len(cookies, 1)
	s.Equal(HankoThirdpartyStateCookie, cookies[0].Name)
	s.Equal(-1, cookies[0].MaxAge)
}
