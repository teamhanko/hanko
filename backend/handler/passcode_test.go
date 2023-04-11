package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"gopkg.in/gomail.v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewPasscodeHandler(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(&config.Config{}, test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{}, test.NewAuditLogger())
	assert.NoError(t, err)
	assert.NotEmpty(t, passcodeHandler)
}

func TestPasscodeHandler_Init(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(&config.Config{}, test.NewPersister(users, nil, nil, nil, nil, nil, nil, emails, nil, nil, nil), sessionManager{}, mailer{}, test.NewAuditLogger())
	require.NoError(t, err)

	body := dto.PasscodeInitRequest{
		UserId: userId,
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/initialize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, passcodeHandler.Init(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	}
}

func TestPasscodeHandler_Init_UnknownUserId(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(&config.Config{}, test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{}, test.NewAuditLogger())
	require.NoError(t, err)

	body := dto.PasscodeInitRequest{
		UserId: "04603148-036d-403b-bf34-cfe237974ef9",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/initialize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = passcodeHandler.Init(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestPasscodeHandler_Finish(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(&config.Config{}, test.NewPersister(users, passcodes(), nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{}, test.NewAuditLogger())
	require.NoError(t, err)

	body := dto.PasscodeFinishRequest{
		Id:   "08ee61aa-0946-4ecf-a8bd-e14c604329e2",
		Code: "123456",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, passcodeHandler.Finish(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	}
}

func TestPasscodeHandler_Finish_WrongCode(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(&config.Config{}, test.NewPersister(nil, passcodes(), nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{}, test.NewAuditLogger())
	require.NoError(t, err)

	body := dto.PasscodeFinishRequest{
		Id:   "08ee61aa-0946-4ecf-a8bd-e14c604329e2",
		Code: "012345",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = passcodeHandler.Finish(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	}
}

func TestPasscodeHandler_Finish_WrongCode_3_Times(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(&config.Config{}, test.NewPersister(nil, passcodes(), nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{}, test.NewAuditLogger())
	require.NoError(t, err)

	body := dto.PasscodeFinishRequest{
		Id:   "08ee61aa-0946-4ecf-a8bd-e14c604329e2",
		Code: "012345",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = passcodeHandler.Finish(c)
		if i < 2 {
			if assert.Error(t, err) {
				httpError := dto.ToHttpError(err)
				assert.Equal(t, http.StatusUnauthorized, httpError.Code)
			}
		} else {
			if assert.Error(t, err) {
				httpError := dto.ToHttpError(err)
				assert.Equal(t, http.StatusGone, httpError.Code)
			}
		}
	}
}

func TestPasscodeHandler_Finish_WrongId(t *testing.T) {
	passcodeHandler, err := NewPasscodeHandler(&config.Config{}, test.NewPersister(nil, passcodes(), nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, mailer{}, test.NewAuditLogger())
	require.NoError(t, err)

	body := dto.PasscodeFinishRequest{
		Id:   "1bc9a074-577d-497e-87da-8eaf50f32a26",
		Code: "123456",
	}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = passcodeHandler.Finish(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
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
