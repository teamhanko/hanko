package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(userSuite))
}

type userSuite struct {
	test.Suite
}

func (s *userSuite) TestUserHandler_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	body := UserCreateBody{Email: "jane.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		s.NoError(err)
		s.False(user.ID.IsNil())

		count, err := s.Storage.GetUserPersister().Count(uuid.Nil, "")
		s.NoError(err)
		s.Equal(1, count)

		email, err := s.Storage.GetEmailPersister().FindByAddress(body.Email)
		s.NoError(err)
		s.NotNil(email)
	}
}

func (s *userSuite) TestUserHandler_Create_CaseInsensitive() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	body := UserCreateBody{Email: "JANE.DOE@EXAMPLE.COM"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		s.NoError(err)
		s.False(user.ID.IsNil())

		count, err := s.Storage.GetUserPersister().Count(uuid.Nil, "")
		s.NoError(err)
		s.Equal(1, count)

		email, err := s.Storage.GetEmailPersister().FindByAddress(strings.ToLower(body.Email))
		s.NoError(err)
		s.NotNil(email)
	}
}

func (s *userSuite) TestUserHandler_Create_UserExists() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user")
	s.Require().NoError(err)

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	body := UserCreateBody{Email: "john.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusConflict, rec.Code) {
		httpError := dto.HTTPError{}
		err := json.Unmarshal(rec.Body.Bytes(), &httpError)
		s.NoError(err)
		s.Equal(http.StatusConflict, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Create_UserExists_CaseInsensitive() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user")
	s.Require().NoError(err)

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	body := UserCreateBody{Email: "JOHN.DOE@EXAMPLE.COM"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusConflict, rec.Code) {
		httpError := dto.HTTPError{}
		err := json.Unmarshal(rec.Body.Bytes(), &httpError)
		s.NoError(err)
		s.Equal(http.StatusConflict, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Create_InvalidEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"email": 123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusBadRequest, rec.Code) {
		httpError := dto.HTTPError{}
		err := json.Unmarshal(rec.Body.Bytes(), &httpError)
		s.NoError(err)
		s.Equal(http.StatusBadRequest, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Create_EmailMissing() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"bogus": 123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusBadRequest, rec.Code) {
		httpError := dto.HTTPError{}
		err := json.Unmarshal(rec.Body.Bytes(), &httpError)
		s.NoError(err)
		s.Equal(http.StatusBadRequest, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig.Session)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}
	token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil(userId))
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", userId), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		s.Equal(rec.Code, http.StatusOK)
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		s.NoError(err)
		s.Equal(userId, user.ID.String())
		s.Equal(len(user.WebauthnCredentials), 0)
	}
}

func (s *userSuite) TestUserHandler_GetUserWithWebAuthnCredential() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig.Session)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}
	token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil(userId))
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", userId), nil)
	rec := httptest.NewRecorder()
	req.AddCookie(cookie)

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		s.Equal(rec.Code, http.StatusOK)
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		s.Require().NoError(err)
		s.Equal(userId, user.ID.String())
		s.Equal(len(user.WebauthnCredentials), 1)
	}
}

func (s *userSuite) TestUserHandler_Get_InvalidUserId() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig.Session)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}
	token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil(userId))
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/users/invalidUserId", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusForbidden, rec.Code) {
		httpError := dto.HTTPError{}
		err := json.Unmarshal(rec.Body.Bytes(), &httpError)
		s.Require().NoError(err)
		s.Equal(http.StatusForbidden, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail_InvalidEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusBadRequest, rec.Code) {
		httpError := dto.HTTPError{}
		err := json.Unmarshal(rec.Body.Bytes(), &httpError)
		s.Require().NoError(err)
		s.Equal(http.StatusBadRequest, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail_InvalidJson() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`"email": "123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail_UserNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "unknownAddress@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNotFound, rec.Code)
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "john.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		response := struct {
			UserId   string `json:"id"`
			Verified bool   `json:"verified"`
		}{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		s.NoError(err)
		s.Equal(userId, response.UserId)
		s.Equal(true, response.Verified)
	}
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail_CaseInsensitive() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "JOHN.DOE@EXAMPLE.COM"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		response := struct {
			UserId   string `json:"id"`
			Verified bool   `json:"verified"`
		}{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		s.NoError(err)
		s.Equal(userId, response.UserId)
		s.Equal(true, response.Verified)
	}
}

func (s *userSuite) TestUserHandler_Me() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig.Session)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}
	token, err := sessionManager.GenerateJWT(uuid.FromStringOrNil(userId))
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {

		response := struct {
			UserId string `json:"id"`
		}{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		s.NoError(err)
		s.Equal(userId, response.UserId)
	}
}

func (s *userSuite) TestUserHandler_Logout() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	userId, _ := uuid.NewV4()
	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil)

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	if err != nil {
		panic(fmt.Errorf("failed to create jwk manager: %w", err))
	}
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig.Session)
	if err != nil {
		panic(fmt.Errorf("failed to create session generator: %w", err))
	}
	token, err := sessionManager.GenerateJWT(userId)
	s.Require().NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusNoContent, rec.Code) {
		cookie := rec.Header().Get("Set-Cookie")
		s.NotEmpty(cookie)

		split := strings.Split(cookie, ";")
		s.Equal("Max-Age=0", strings.TrimSpace(split[2]))
	}
}
