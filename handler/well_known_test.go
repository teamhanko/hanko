package handler

import (
	"errors"
	"github.com/labstack/echo/v4"
	jwk2 "github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/crypto/jwk"
	"github.com/teamhanko/hanko/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type faultyJwkManager struct {
}

func (f faultyJwkManager) GenerateKey() (jwk2.Key, error) {
	panic("implement me")
}

func (f faultyJwkManager) GetPublicKeys() ([]jwk2.Key, error) {
	return nil, errors.New("No Public Keys!")
}

func (f faultyJwkManager) GetSigningKey() (jwk2.Key, error) {
	panic("implement me")
}

func TestSomethingWrongWithKeys(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	jwkMan := faultyJwkManager{}
	h, err := NewWellKnownHandler(jwkMan)
	assert.NoError(t, err)

	if assert.NoError(t, h.GetPublicKeys(c)) {
		assert.Equal(t, `{"Code":500,"Message":"Internal Server Error"}`, strings.TrimSuffix(rec.Body.String(), "\n"))
		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
	}
}

func TestGetPublicKeys(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	jwkMan, err := jwk.NewDefaultManager([]string{"superRandomAndSecure"}, test.NewJwkPersister(nil))
	assert.NoError(t, err)
	h, err := NewWellKnownHandler(jwkMan)
	assert.NoError(t, err)

	if assert.NoError(t, h.GetPublicKeys(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	}
}
