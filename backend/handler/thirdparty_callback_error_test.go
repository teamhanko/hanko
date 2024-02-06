package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/h2non/gock"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"github.com/teamhanko/hanko/backend/utils"
)

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_SignUpUserConflict() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New(thirdparty.GoogleOauthTokenEndpoint).
		Post("/").
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New(thirdparty.GoogleUserInfoEndpoint).
		Get("/").
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_email_already_exists",
			Email:         "test-no-identity@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", stateResponse.EncryptedState), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(stateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeUserConflict, location.Query().Get("error"))
		s.Equal("third party account linking for existing user with same email disallowed", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_SignInMultipleAccounts() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/thirdparty")
	s.NoError(err)

	gock.New(thirdparty.GoogleOauthTokenEndpoint).
		Post("/").
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New(thirdparty.GoogleUserInfoEndpoint).
		Get("/").
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "provider-primary-email-changed@example.com",
			EmailVerified: true,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", stateResponse.EncryptedState), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(stateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeMultipleAccounts, location.Query().Get("error"))
		s.Equal(fmt.Sprintf("cannot identify associated user: '%s' is used by multiple accounts", "provider-primary-email-changed@example.com"), location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
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

		s.Equal(thirdparty.ErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("State is a required field", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_StateMismatch() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	mismatchedstateResponse, err := thirdparty.GenerateState(cfg, "github", "https://foo.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", stateResponse.EncryptedState), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(mismatchedstateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("could not verify state", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_NoThirdPartyCookie() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", stateResponse.EncryptedState), nil)

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("thirdparty state cookie is missing", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_ProviderError() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	providerError := "access_denied"
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s&error=%s", stateResponse.EncryptedState, providerError), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(stateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(providerError, location.Query().Get("error"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_ProviderDisabled() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"github"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", stateResponse.EncryptedState), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(stateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("google provider is disabled", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_NoAuthCode() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?state=%s", stateResponse.EncryptedState), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(stateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("auth code missing from request", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_OAuthTokenExchange() {
	defer gock.Off()
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	gock.New(thirdparty.GoogleOauthTokenEndpoint).
		Post("/").
		Reply(400).
		JSON(map[string]string{"error": "incorrect_client_credentials"})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", stateResponse.EncryptedState), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(stateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeInvalidRequest, location.Query().Get("error"))
		s.Equal("could not exchange authorization code for access token", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}

func (s *thirdPartySuite) TestThirdPartyHandler_Callback_Error_VerificationRequiredUnverifiedProviderEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	gock.New(thirdparty.GoogleOauthTokenEndpoint).
		Post("/").
		Reply(200).
		JSON(map[string]string{"access_token": "fakeAccessToken"})

	gock.New(thirdparty.GoogleUserInfoEndpoint).
		Get("/").
		Reply(200).
		JSON(&thirdparty.GoogleUser{
			ID:            "google_abcde",
			Email:         "test-google-signup@example.com",
			EmailVerified: false,
		})

	cfg := s.setUpConfig([]string{"google"}, []string{"https://example.com"})
	cfg.Emails.RequireVerification = true

	stateResponse, err := thirdparty.GenerateState(cfg, "google", "https://example.com")
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/thirdparty/callback?code=abcde&state=%s", stateResponse.EncryptedState), nil)
	req.AddCookie(&http.Cookie{
		Name:  utils.HankoThirdpartyStateCookie,
		Value: string(stateResponse.EncryptedState),
	})

	c, rec := s.setUpContext(req)
	handler := s.setUpHandler(cfg)

	if s.NoError(handler.Callback(c)) {
		s.Equal(http.StatusTemporaryRedirect, rec.Code)
		location, err := rec.Result().Location()
		s.NoError(err)

		s.Equal(thirdparty.ErrorCodeUnverifiedProviderEmail, location.Query().Get("error"))
		s.Equal("third party provider email must be verified", location.Query().Get("error_description"))

		logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"thirdparty_signin_signup_failed"}, "", "", "", "")
		s.NoError(lerr)
		s.Len(logs, 1)
	}
}
