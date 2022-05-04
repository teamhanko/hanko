package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
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

	if assert.NoError(t, handler.Create(c)) {
		assert.Equal(t, http.StatusConflict, rec.Code)
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

	if assert.NoError(t, handler.Create(c)) {
		assert.Equal(t, rec.Code, http.StatusBadRequest)
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

	if assert.NoError(t, handler.Create(c)) {
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		apiError := dto.ApiError{}
		err := json.Unmarshal(rec.Body.Bytes(), &apiError)
		require.NoError(t, err)
		assert.Equal(t, 1, len(apiError.ValidationErrors))
	}
}
