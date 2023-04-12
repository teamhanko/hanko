package handler

import (
	"fmt"
	"github.com/h2non/gock"
	"github.com/teamhanko/hanko/backend/test"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
