package handler

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestUserHandlerAdmin_Delete(t *testing.T) {
	userId, _ := uuid.NewV4()
	users := []models.User{
		{
			ID:        userId,
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

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	err := handler.Delete(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
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

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	err := handler.Delete(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusNotFound, httpError.Code)
	}
}

func TestUserHandlerAdmin_List(t *testing.T) {
	users := []models.User{
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var users []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, "2", rec.Header().Get("X-Total-Count"))
		assert.Equal(t, "<http://example.com/users?page=1&per_page=20>; rel=\"first\"", rec.Header().Get("Link"))
	}
}

func TestUserHandlerAdmin_List_Pagination(t *testing.T) {
	users := []models.User{
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
		func() models.User {
			userId, _ := uuid.NewV4()
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "1")
	q.Set("page", "2")
	req := httptest.NewRequest(http.MethodGet, "/users?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var got []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(got))
		assert.Equal(t, "2", rec.Header().Get("X-Total-Count"))
		assert.Equal(t, "<http://example.com/users?page=1&per_page=1>; rel=\"first\",<http://example.com/users?page=1&per_page=1>; rel=\"prev\"", rec.Header().Get("Link"))
	}
}

func TestUserHandlerAdmin_List_NoUsers(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "1")
	q.Set("page", "1")
	req := httptest.NewRequest(http.MethodGet, "/users?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	if assert.NoError(t, handler.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var got []models.User
		err := json.Unmarshal(rec.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(got))
		assert.Equal(t, "0", rec.Header().Get("X-Total-Count"))
		assert.Equal(t, "<http://example.com/users?page=1&per_page=1>; rel=\"first\"", rec.Header().Get("Link"))
	}
}

func TestUserHandlerAdmin_List_InvalidPaginationParam(t *testing.T) {
	e := echo.New()

	q := make(url.Values)
	q.Set("per_page", "invalidPerPageValue")
	req := httptest.NewRequest(http.MethodGet, "/users?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewUserHandlerAdmin(p)

	err := handler.List(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}
