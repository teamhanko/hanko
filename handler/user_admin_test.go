package handler

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/persistence/models"
	"github.com/teamhanko/hanko/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUserHandlerAdmin_Delete(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Delete(c)) {
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}
}

func TestUserHandlerAdmin_Delete_InvalidUserId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues("invalidId")

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Delete(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUserHandlerAdmin_Delete_UnknownUserId(t *testing.T) {
	userId, _ := uuid.NewV4()
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Delete(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestUserHandlerAdmin_Patch(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "jane.doe@example.com", "verified": true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(userId.String())

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Patch(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestUserHandlerAdmin_Patch_InvalidUserIdAndEmail(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "invalidEmail"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues("invalidUserId")

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Patch(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		apiError := dto.ApiError{}
		err := json.Unmarshal(rec.Body.Bytes(), &apiError)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(apiError.ValidationErrors))
	}
}

func TestUserHandlerAdmin_Patch_EmailNotAvailable(t *testing.T) {
	users := []models.User{
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "john.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				Email:     "jane.doe@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "jane.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues(users[0].ID.String())

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Patch(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUserHandlerAdmin_Patch_UnknownUserId(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"email": "jane.doe@example.com", "verified": true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	unknownUserId, _ := uuid.NewV4()
	c.SetParamValues(unknownUserId.String())

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Patch(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestUserHandlerAdmin_Patch_InvalidJson(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
			Email:     "john.doe@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`"email: "jane.doe@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	unknownUserId, _ := uuid.NewV4()
	c.SetParamValues(unknownUserId.String())

	p := test.NewPersister(users, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.Patch(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}
