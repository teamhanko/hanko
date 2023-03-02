package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
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
	suite.Suite
	storage persistence.Storage
	db      *test.TestDB
}

func (s *userSuite) SetupSuite() {
	if testing.Short() {
		return
	}
	dialect := "postgres"
	db, err := test.StartDB("user_test", dialect)
	s.NoError(err)
	storage, err := persistence.New(config.Database{
		Url: db.DatabaseUrl,
	})
	s.NoError(err)

	s.storage = storage
	s.db = db
}

func (s *userSuite) SetupTest() {
	if s.db != nil {
		err := s.storage.MigrateUp()
		s.NoError(err)
	}
}

func (s *userSuite) TearDownTest() {
	if s.db != nil {
		err := s.storage.MigrateDown(-1)
		s.NoError(err)
	}
}

func (s *userSuite) TearDownSuite() {
	if s.db != nil {
		s.NoError(test.PurgeDB(s.db))
	}
}

func (s *userSuite) TestUserHandler_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	body := UserCreateBody{Email: "jane.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	if s.NoError(handler.Create(c)) {
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		s.NoError(err)
		s.False(user.ID.IsNil())

		count, err := s.storage.GetUserPersister().Count(uuid.Nil, "")
		s.NoError(err)
		s.Equal(1, count)

		email, err := s.storage.GetEmailPersister().FindByAddress(body.Email)
		s.NoError(err)
		s.NotNil(email)
	}
}

func (s *userSuite) TestUserHandler_Create_CaseInsensitive() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	body := UserCreateBody{Email: "JANE.DOE@EXAMPLE.COM"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	if s.NoError(handler.Create(c)) {
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		s.NoError(err)
		s.False(user.ID.IsNil())

		count, err := s.storage.GetUserPersister().Count(uuid.Nil, "")
		s.NoError(err)
		s.Equal(1, count)

		email, err := s.storage.GetEmailPersister().FindByAddress(strings.ToLower(body.Email))
		s.NoError(err)
		s.NotNil(email)
	}
}

func (s *userSuite) TestUserHandler_Create_UserExists() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/user")
	s.Require().NoError(err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := UserCreateBody{Email: "john.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	err = handler.Create(c)
	if s.Error(err) {
		httpError := dto.ToHttpError(err)
		s.Equal(http.StatusConflict, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Create_UserExists_CaseInsensitive() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/user")
	s.Require().NoError(err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := UserCreateBody{Email: "JOHN.DOE@EXAMPLE.COM"}
	bodyJson, err := json.Marshal(body)
	s.NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	err = handler.Create(c)
	if s.Error(err) {
		httpError := dto.ToHttpError(err)
		s.Equal(http.StatusConflict, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Create_InvalidEmail() {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"email": 123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, nil, nil, nil)

	err := handler.Create(c)
	if s.Error(err) {
		httpError := dto.ToHttpError(err)
		s.Equal(http.StatusBadRequest, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Create_EmailMissing() {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"bogus": 123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, nil, nil, nil)

	err := handler.Create(c)
	if s.Error(err) {
		httpError := dto.ToHttpError(err)
		s.Equal(http.StatusBadRequest, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/user")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId)

	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId)
	s.Require().NoError(err)
	c.Set("session", token)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	if s.NoError(handler.Get(c)) {
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
	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId)

	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId)
	s.Require().NoError(err)
	c.Set("session", token)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	if s.NoError(handler.Get(c)) {
		s.Equal(rec.Code, http.StatusOK)
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		s.Require().NoError(err)
		s.Equal(userId, user.ID.String())
		s.Equal(len(user.WebauthnCredentials), 1)
	}
}

func (s *userSuite) TestUserHandler_Get_InvalidUserId() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/invalidUserId", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, "completelyDifferentUserId")
	s.Require().NoError(err)
	c.Set("session", token)

	handler := NewUserHandler(&defaultConfig, nil, sessionManager{}, test.NewAuditLogger())

	err = handler.Get(c)
	if s.Error(err) {
		httpError := dto.ToHttpError(err)
		s.Equal(http.StatusForbidden, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail_InvalidEmail() {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, nil, nil, nil)

	err := handler.GetUserIdByEmail(c)
	if s.Error(err) {
		httpError := dto.ToHttpError(err)
		s.Equal(http.StatusBadRequest, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail_InvalidJson() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`"email": "123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, nil, nil, nil)

	s.Error(handler.GetUserIdByEmail(c))
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail_UserNotFound() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "unknownAddress@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	err := handler.GetUserIdByEmail(c)
	if s.Error(err) {
		httpError := dto.ToHttpError(err)
		s.Equal(http.StatusNotFound, httpError.Code)
	}
}

func (s *userSuite) TestUserHandler_GetUserIdByEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "john.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	if s.NoError(handler.GetUserIdByEmail(c)) {
		s.Equal(http.StatusOK, rec.Code)
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
	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "JOHN.DOE@EXAMPLE.COM"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	if s.NoError(handler.GetUserIdByEmail(c)) {
		s.Equal(http.StatusOK, rec.Code)
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
	err := test.LoadFixtures(s.db.DbCon, s.db.Dialect, "../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId)
	s.Require().NoError(err)
	c.Set("session", token)

	handler := NewUserHandler(&defaultConfig, s.storage, sessionManager{}, test.NewAuditLogger())

	if s.NoError(handler.Me(c)) {
		s.Equal(http.StatusOK, rec.Code)
		response := struct {
			UserId string `json:"id"`
		}{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		s.NoError(err)
		s.Equal(userId, response.UserId)
	}
}

func TestUserHandler_Logout(t *testing.T) {
	userId, _ := uuid.NewV4()

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(&defaultConfig, p, sessionManager{}, test.NewAuditLogger())

	if assert.NoError(t, handler.Logout(c)) {
		assert.Equal(t, http.StatusNoContent, rec.Code)
		cookie := rec.Header().Get("Set-Cookie")
		assert.NotEmpty(t, cookie)

		split := strings.Split(cookie, ";")
		assert.Equal(t, "Max-Age=0", strings.TrimSpace(split[1]))
	}
}
