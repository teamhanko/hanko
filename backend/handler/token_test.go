package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(tokenSuite))
}

type tokenSuite struct {
	test.Suite
}

func (s *tokenSuite) TestToken_Validate_TokenInCookie() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/token")

	// must create and insert a valid token manually instead of using fixtures, because token
	// validation is time sensitive (expiration is checked relative to current time)
	uId := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")
	token, err := models.NewToken(uId)
	s.NoError(err)
	err = s.Storage.GetTokenPersister().Create(*token)
	s.NoError(err)

	body := TokenValidationBody{Value: token.Value}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	cfg := s.setupConfig()
	cfg.Session.EnableAuthTokenHeader = false
	e := NewPublicRouter(cfg, s.Storage, nil, nil)
	e.ServeHTTP(rec, req)

	s.Equal(rec.Code, http.StatusOK)
	t, err := s.Storage.GetTokenPersister().GetByValue(token.Value)
	s.NoError(err)
	s.Nil(t)

	s.Empty(rec.Header().Get("X-Auth-Token"))
	cookies := rec.Result().Cookies()
	rec.Result().Cookies()
	s.NotEmpty(cookies)
	for _, cookie := range cookies {
		if cookie.Name == "hanko" {
			s.Regexp(".*\\..*\\..*", cookie.Value)
		}
	}

	logs, err := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"token_exchange_succeeded"}, "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5", "", "", "")
	s.Len(logs, 1)
}

func (s *tokenSuite) TestToken_Validate_TokenInHeader() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/token")

	// must create and insert a valid token manually instead of using fixtures, because token
	// validation is time sensitive (expiration is checked relative to current time)
	uId := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")
	token, err := models.NewToken(uId)
	s.NoError(err)
	err = s.Storage.GetTokenPersister().Create(*token)
	s.NoError(err)

	body := TokenValidationBody{Value: token.Value}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	cfg := s.setupConfig()
	e := NewPublicRouter(cfg, s.Storage, nil, nil)
	e.ServeHTTP(rec, req)

	s.Equal(rec.Code, http.StatusOK)
	t, err := s.Storage.GetTokenPersister().GetByValue(token.Value)
	s.NoError(err)
	s.Nil(t)

	s.Empty(rec.Result().Cookies())
	responseToken := rec.Header().Get("X-Auth-Token")
	s.NotEmpty(responseToken)
	s.Regexp(".*\\..*\\..*", responseToken)

	logs, err := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"token_exchange_succeeded"}, "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5", "", "", "")
	s.Len(logs, 1)
}

func (s *tokenSuite) TestToken_Validate_ExpiredToken() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/token")

	expiredTokenValue := "Trkauhl3q7XVxw5JcDH80lTe1KxzydIw0OcizH7umWk="
	body := TokenValidationBody{Value: expiredTokenValue}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := NewPublicRouter(s.setupConfig(), s.Storage, nil, nil)
	e.ServeHTTP(rec, req)

	s.Equal(rec.Code, http.StatusUnprocessableEntity)
	var errorResponse echo.HTTPError
	marshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	s.NoError(marshalErr)
	s.Contains(errorResponse.Message, "expired")

	logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"token_exchange_failed"}, "", "", "", "")
	s.NoError(lerr)
	s.Len(logs, 1)
}

func (s *tokenSuite) TestToken_Validate_MissingTokenFromRequest() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	req := httptest.NewRequest(http.MethodPost, "/token", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := NewPublicRouter(s.setupConfig(), s.Storage, nil, nil)
	e.ServeHTTP(rec, req)

	s.Equal(rec.Code, http.StatusBadRequest)
	var errorResponse echo.HTTPError
	marshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	s.NoError(marshalErr)
	s.Contains("value is a required field", errorResponse.Message)

	logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"token_exchange_failed"}, "", "", "", "")
	s.NoError(lerr)
	s.Len(logs, 1)
}

func (s *tokenSuite) TestToken_Validate_InvalidJson() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := NewPublicRouter(s.setupConfig(), s.Storage, nil, nil)
	e.ServeHTTP(rec, req)

	s.Equal(rec.Code, http.StatusBadRequest)

	logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"token_exchange_failed"}, "", "", "", "")
	s.NoError(lerr)
	s.Len(logs, 1)

}

func (s *tokenSuite) TestToken_Validate_TokenNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	uId := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")
	token, err := models.NewToken(uId)
	s.NoError(err)

	body := TokenValidationBody{Value: token.Value}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e := NewPublicRouter(s.setupConfig(), s.Storage, nil, nil)
	e.ServeHTTP(rec, req)

	s.Equal(rec.Code, http.StatusNotFound)
	var errorResponse echo.HTTPError
	marshalErr := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	s.NoError(marshalErr)
	s.Contains("token not found", errorResponse.Message)

	logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{"token_exchange_failed"}, "", "", "", "")
	s.NoError(lerr)
	s.Len(logs, 1)
}

func (s *tokenSuite) setupConfig() *config.Config {
	cfg := test.DefaultConfig
	cfg.Session.EnableAuthTokenHeader = true
	cfg.AuditLog.Storage.Enabled = true
	return &cfg
}
