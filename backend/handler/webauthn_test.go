package handler

import (
	"encoding/base64"
	"encoding/json"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebauthnSuite(t *testing.T) {
	suite.Run(t, new(webauthnSuite))
}

type webauthnSuite struct {
	test.Suite
}

func (s *webauthnSuite) TestWebauthnHandler_NewHandler() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}
	handler, err := NewWebauthnHandler(&test.DefaultConfig, s.Storage, s.GetDefaultSessionManager(), test.NewAuditLogger())
	s.NoError(err)
	s.NotEmpty(handler)
}

func (s *webauthnSuite) TestWebauthnHandler_BeginRegistration() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../test/fixtures/webauthn")
	s.Require().NoError(err)

	userId := "ec4ef049-5b88-4321-a173-21b0eff06a04"

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	sessionManager := s.GetDefaultSessionManager()
	token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil(userId))
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/webauthn/registration/initialize", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		creationOptions := protocol.CredentialCreation{}
		err = json.Unmarshal(rec.Body.Bytes(), &creationOptions)
		s.NoError(err)

		uId, err := base64.RawURLEncoding.DecodeString(creationOptions.Response.User.ID.(string))
		s.Require().NoError(err)

		s.NotEmpty(creationOptions.Response.Challenge)
		s.Equal(uuid.FromStringOrNil(userId).Bytes(), uId)
		s.Equal(test.DefaultConfig.Webauthn.RelyingParty.Id, creationOptions.Response.RelyingParty.ID)
		s.Equal(protocol.ResidentKeyRequirementRequired, creationOptions.Response.AuthenticatorSelection.ResidentKey)
		s.Equal(protocol.VerificationRequired, creationOptions.Response.AuthenticatorSelection.UserVerification)
		s.True(*creationOptions.Response.AuthenticatorSelection.RequireResidentKey)
	}
}

func (s *webauthnSuite) TestWebauthnHandler_FinalizeRegistration() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../test/fixtures/webauthn_registration")
	s.Require().NoError(err)

	userId := "ec4ef049-5b88-4321-a173-21b0eff06a04"

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	sessionManager := s.GetDefaultSessionManager()
	token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil(userId))
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)

	body := `{
"id": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"rawId": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"type": "public-key",
"response": {
"attestationObject": "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVjeSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFYmehnq3OAAI1vMYKZIsLJfHwVQMAWgGhXZHA-Erj4xfo8FKEcB_PmR7mOUVuOn7GZhLwV-kTSh2hrVc6QE7NOikFYXiDo2M_mJ3huHJkDnnc5dHtIxfedbpMdex5fY3hoFs-fwymQjtdqdvti5c4x6UBAgMmIAEhWCDxvVrRgK4vpnr6JxTx-KfpSNyQUtvc47ryryZmj-P5kSJYIDox8N9bHQBrxN-b5kXqfmj3GwAJW7nNCh8UPbus3B6I",
"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoidE9yTkRDRDJ4UWY0ekZqRWp3eGFQOGZPRXJQM3p6MDhyTW9UbEpHdG5LVSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0"
}
}`

	req := httptest.NewRequest(http.MethodPost, "/webauthn/registration/finalize", strings.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		s.Equal(`{"credential_id":"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH","user_id":"ec4ef049-5b88-4321-a173-21b0eff06a04"}`, strings.TrimSpace(rec.Body.String()))
	}

	req2 := httptest.NewRequest(http.MethodPost, "/webauthn/registration/finalize", strings.NewReader(body))
	req2.AddCookie(cookie)
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)
	s.Equal(http.StatusBadRequest, rec2.Code)
}

func (s *webauthnSuite) TestWebauthnHandler_BeginAuthentication() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../test/fixtures/webauthn")
	s.Require().NoError(err)

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)
	req := httptest.NewRequest(http.MethodPost, "/webauthn/login/initialize", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		assertionOptions := protocol.CredentialAssertion{}
		err := json.Unmarshal(rec.Body.Bytes(), &assertionOptions)
		s.Require().NoError(err)
		s.NotEmpty(assertionOptions.Response.Challenge)
		s.Equal(assertionOptions.Response.UserVerification, protocol.VerificationRequired)
		s.Equal(test.DefaultConfig.Webauthn.RelyingParty.Id, assertionOptions.Response.RelyingPartyID)
	}
}

func (s *webauthnSuite) TestWebauthnHandler_FinalizeAuthentication() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../test/fixtures/webauthn")
	s.Require().NoError(err)

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	body := `{
"id": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"rawId": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"type": "public-key",
"response": {
"authenticatorData": "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFYmezOw",
"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiZ0tKS21oOTB2T3BZTzU1b0hwcWFIWF9vTUNxNG9UWnQtRDBiNnRlSXpyRSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
"signature": "MEYCIQDi2vYVspG6pf38I4GyQCPOojGbvX4nwSPXCi0hm80twAIhAO3EWjhAnj0UpjU_l0AH5sEh3zq4LDvkvo3AUqaqfGYD",
"userHandle": "7E7wSVuIQyGhcyGw7_BqBA"
}
}`

	req := httptest.NewRequest(http.MethodPost, "/webauthn/login/finalize", strings.NewReader(body))
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		cookies := rec.Result().Cookies()
		if s.NotEmpty(cookies) {
			for _, cookie := range cookies {
				if cookie.Name == "hanko" {
					s.Regexp(".*\\..*\\..*", cookie.Value) // check if cookie contains a jwt
				}
			}
		}
		s.Equal(`{"credential_id":"AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH","user_id":"ec4ef049-5b88-4321-a173-21b0eff06a04"}`, strings.TrimSpace(rec.Body.String()))
	}

	req2 := httptest.NewRequest(http.MethodPost, "/webauthn/login/finalize", strings.NewReader(body))
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	if s.Equal(http.StatusUnauthorized, rec2.Code) {
		httpError := echo.HTTPError{}
		err = json.Unmarshal(rec2.Body.Bytes(), &httpError)
		s.NoError(err)
		s.Equal("Stored challenge and received challenge do not match", httpError.Message)
	}
}

func (s *webauthnSuite) GetDefaultSessionManager() session.Manager {
	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	s.Require().NoError(err)
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig)
	s.Require().NoError(err)

	return sessionManager
}

var userId = "ec4ef049-5b88-4321-a173-21b0eff06a04"

var defaultConfig = config.Config{
	Webauthn: config.WebauthnSettings{
		RelyingParty: config.RelyingParty{
			Id:          "localhost",
			DisplayName: "Test Relying Party",
			Icon:        "",
			Origins:     []string{"http://localhost:8080"},
		},
		Timeout: 60000,
	},
	Secrets: config.Secrets{
		Keys: []string{"abcdefghijklmnop"},
	},
	Passcode: config.Passcode{Smtp: config.SMTP{
		Host: "localhost",
		Port: "2500",
	}},
}

type sessionManager struct {
}

func (s sessionManager) GenerateJWT(uuid uuid.UUID) (string, error) {
	return userId, nil
}

func (s sessionManager) GenerateCookie(token string) (*http.Cookie, error) {
	return &http.Cookie{
		Name:     "hanko",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func (s sessionManager) DeleteCookie() (*http.Cookie, error) {
	return &http.Cookie{
		Name:     "hanko",
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}, nil
}

func (s sessionManager) Verify(token string) (jwt.Token, error) {
	return nil, nil
}

var uId, _ = uuid.FromString(userId)

var emails = []models.Email{
	{
		ID:           uId,
		Address:      "john.doe@example.com",
		PrimaryEmail: &models.PrimaryEmail{ID: uId},
	},
}

var users = []models.User{
	func() models.User {
		return models.User{
			ID:        uId,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Emails:    emails,
		}
	}(),
}
