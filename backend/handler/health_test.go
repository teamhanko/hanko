package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler_Ready(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := NewHealthHandler()

	if assert.NoError(t, h.Ready(c)) {
		assert.Equal(t, `{"ready":true}`, strings.TrimSuffix(rec.Body.String(), "\n"))
	}
}

func TestHealthHandler_Alive(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health/alive", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := NewHealthHandler()

	if assert.NoError(t, h.Alive(c)) {
		assert.Equal(t, `{"alive":true}`, strings.TrimSuffix(rec.Body.String(), "\n"))
	}
}
