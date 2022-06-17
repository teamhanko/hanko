package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/persistence/models"
	"github.com/teamhanko/hanko/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUserHandler_Create(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	body := UserCreateBody{Email: "jane.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	if assert.NoError(t, handler.Create(c)) {
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.False(t, user.ID.IsNil())
		assert.Equal(t, body.Email, user.Email)
	}
}

func TestUserHandler_Create_UserExists(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := UserCreateBody{Email: "john.doe@example.com"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	err = handler.Create(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusConflict, httpError.Code)
	}
}

func TestUserHandler_Create_InvalidEmail(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"email": 123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	err := handler.Create(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandler_Create_EmailMissing(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"bogus": 123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	err := handler.Create(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandler_Get(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	if assert.NoError(t, handler.Get(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, userId, user.ID)
		assert.Equal(t, len(user.WebauthnCredentials), 0)
	}
}

func TestUserHandler_GetUserWithWebAuthnCredential(t *testing.T) {
	userId, _ := uuid.NewV4()
	aaguid, _ := uuid.FromString("adce0002-35bc-c60a-648b-0b25f1f05503")
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			WebauthnCredentials: []models.WebauthnCredential{
				{
					ID:              "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
					UserId:          userId,
					PublicKey:       "pQECAyYgASFYIPG9WtGAri-mevonFPH4p-lI3JBS29zjuvKvJmaP4_mRIlggOjHw31sdAGvE35vmRep-aPcbAAlbuc0KHxQ9u6zcHog",
					AttestationType: "none",
					AAGUID:          aaguid,
					SignCount:       1650958750,
					CreatedAt:       time.Time{},
					UpdatedAt:       time.Time{},
				},
			},
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	if assert.NoError(t, handler.Get(c)) {
		assert.Equal(t, rec.Code, http.StatusOK)
		user := models.User{}
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		require.NoError(t, err)
		assert.Equal(t, userId, user.ID)
		assert.Equal(t, len(user.WebauthnCredentials), 1)
	}
}

func TestUserHandler_Get_InvalidUserId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/invalidUserId", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, "completelyDifferentUserId")
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	err = handler.Get(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusForbidden, httpError.Code)
	}
}

func TestUserHandler_GetUserIdByEmail_InvalidEmail(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	err := handler.GetUserIdByEmail(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestUserHandler_GetUserIdByEmail_InvalidJson(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`"email": "123}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	assert.Error(t, handler.GetUserIdByEmail(c))
}

func TestUserHandler_GetUserIdByEmail_UserNotFound(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "unknownAddress@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	err := handler.GetUserIdByEmail(c)
	if assert.Error(t, err) {
		assert.Equal(t, http.StatusText(http.StatusNotFound), err.Error())
	}
}

func TestUserHandler_GetUserIdByEmail(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Verified:  true,
		},
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(`{"email": "john.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	if assert.NoError(t, handler.GetUserIdByEmail(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response := struct {
			UserId   string `json:"id"`
			Verified bool   `json:"verified"`
		}{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, users[0].ID.String(), response.UserId)
		assert.Equal(t, users[0].Verified, response.Verified)
	}
}

func TestUserHandler_Me(t *testing.T) {
	userId, _ := uuid.NewV4()

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandler(p)

	if assert.NoError(t, handler.Me(c)) {
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, fmt.Sprintf("/users/%s", userId.String()), rec.Header().Get("Location"))
	}
}
