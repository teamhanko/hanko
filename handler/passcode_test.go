package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/persistence/models"
	"github.com/teamhanko/hanko/test"
	"gopkg.in/gomail.v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewPasscodeHandler(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(config.Passcode{}, test.NewPersister(nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{})
	assert.NoError(t, err)
	assert.NotEmpty(t, passcodeHandler)
}

func TestPasscodeHandler_Init(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(config.Passcode{}, test.NewPersister(users, nil, nil, nil, nil, nil), sessionManager{}, mailer{})
	require.NoError(t, err)

	body := passcodeInit{
		UserId: userId,
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/initialize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, passcodeHandler.Init(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	}
}

func TestPasscodeHandler_Init_UnknownUserId(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(config.Passcode{}, test.NewPersister(users, nil, nil, nil, nil, nil), sessionManager{}, mailer{})
	require.NoError(t, err)

	body := passcodeInit{
		UserId: "04603148-036d-403b-bf34-cfe237974ef9",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/initialize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, passcodeHandler.Init(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
	}
}

func TestPasscodeHandler_Finish(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(config.Passcode{}, test.NewPersister(users, passcodes(), nil, nil, nil, nil), sessionManager{}, mailer{})
	require.NoError(t, err)

	body := passcodeFinish{
		Id:   "08ee61aa-0946-4ecf-a8bd-e14c604329e2",
		Code: "123456",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, passcodeHandler.Finish(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	}
}

func TestPasscodeHandler_Finish_WrongCode(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(config.Passcode{}, test.NewPersister(users, passcodes(), nil, nil, nil, nil), sessionManager{}, mailer{})
	require.NoError(t, err)

	body := passcodeFinish{
		Id:   "08ee61aa-0946-4ecf-a8bd-e14c604329e2",
		Code: "012345",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, passcodeHandler.Finish(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
	}
}

func TestPasscodeHandler_Finish_WrongId(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(config.Passcode{}, test.NewPersister(users, passcodes(), nil, nil, nil, nil), sessionManager{}, mailer{})
	require.NoError(t, err)

	body := passcodeFinish{
		Id:   "1bc9a074-577d-497e-87da-8eaf50f32a26",
		Code: "123456",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, passcodeHandler.Finish(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Result().StatusCode)
	}
}

func passcodes() []models.Passcode {
	now := time.Now()
	return []models.Passcode{{
		ID:        uuid.FromStringOrNil("08ee61aa-0946-4ecf-a8bd-e14c604329e2"),
		UserId:    uuid.FromStringOrNil(userId),
		Ttl:       300,
		Code:      "$2a$12$gBPH9jnbXFmwAGwZMSzYkeXx7oOTElzhvHfiDgj.D7G8q4znvHpMK",
		CreatedAt: now,
		UpdatedAt: now,
	}}
}

type mailer struct {
}

func (m mailer) Send(message *gomail.Message) error {
	return nil
}
