package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPasswordHandler_Set_Create(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := PasswordSetBody{UserID: userId.String(), Password: "verybadpassword"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/password", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, []models.PasswordCredential{}, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	if assert.NoError(t, handler.Set(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}

func TestPasswordHandler_Set_Create_PasswordTooShort(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := PasswordSetBody{UserID: userId.String(), Password: "very"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/password", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, []models.PasswordCredential{}, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{Password: config.Password{MinPasswordLength: 8}}, test.NewAuditLogger())

	err = handler.Set(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestPasswordHandler_Set_Create_PasswordTooLong(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := PasswordSetBody{UserID: userId.String(), Password: "thisIsAVeryLongPasswordThatIsUsedToTestIfAnErrorWillBeReturnedForTooLongPasswords"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/password", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, []models.PasswordCredential{}, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{Password: config.Password{MinPasswordLength: 8}}, test.NewAuditLogger())

	err = handler.Set(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestPasswordHandler_Set_Update(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("verybadpassword"), 12)
	assert.NoError(t, err)

	passwords := []models.PasswordCredential{
		func() models.PasswordCredential {
			pId, _ := uuid.NewV4()
			return models.PasswordCredential{
				ID:        pId,
				UserId:    userId,
				Password:  string(hashedPassword),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := PasswordSetBody{UserID: userId.String(), Password: "anotherbadnewpassword"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/password", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, passwords, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	if assert.NoError(t, handler.Set(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestPasswordHandler_Set_UserNotFound(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := PasswordSetBody{UserID: userId.String(), Password: "anotherbadnewpassword"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/password", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister([]models.User{}, nil, nil, nil, nil, []models.PasswordCredential{}, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	err = handler.Set(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	}
}

func TestPasswordHandler_Set_TokenHasWrongSubject(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("verybadpassword"), 12)
	assert.NoError(t, err)

	passwords := []models.PasswordCredential{
		func() models.PasswordCredential {
			pId, _ := uuid.NewV4()
			return models.PasswordCredential{
				ID:        pId,
				UserId:    userId,
				Password:  string(hashedPassword),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := PasswordSetBody{UserID: userId.String(), Password: "anotherbadnewpassword"}
	bodyJson, err := json.Marshal(body)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/password", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	wrongUid, _ := uuid.NewV4()
	err = token.Set(jwt.SubjectKey, wrongUid.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, nil, passwords, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	err = handler.Set(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusForbidden, httpError.Code)
	}
}

func TestPasswordHandler_Set_BadRequestBody(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := `{"user_id_wrong": "123", "password_wrong": "badpassword"}`

	req := httptest.NewRequest(http.MethodPost, "/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	uId, _ := uuid.NewV4()
	err := token.Set(jwt.SubjectKey, uId.String())
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	err = handler.Set(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	}
}

func TestPasswordHandler_Login(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("verybadpassword"), 12)
	assert.NoError(t, err)

	passwords := []models.PasswordCredential{
		func() models.PasswordCredential {
			pId, _ := uuid.NewV4()
			return models.PasswordCredential{
				ID:        pId,
				UserId:    userId,
				Password:  string(hashedPassword),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := `{"user_id": "ec4ef049-5b88-4321-a173-21b0eff06a04", "password": "verybadpassword"}`

	req := httptest.NewRequest(http.MethodPost, "/password/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, passwords, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	if assert.NoError(t, handler.Login(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		cookies := rec.Result().Cookies()
		if assert.NotEmpty(t, cookies) {
			for _, cookie := range cookies {
				if cookie.Name == "hanko" {
					assert.Equal(t, "ec4ef049-5b88-4321-a173-21b0eff06a04", cookie.Value)
				}
			}
		}
	}
}

func TestPasswordHandler_Login_WrongPassword(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
	users := []models.User{
		func() models.User {
			return models.User{
				ID:        userId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("verybadpassword"), 12)
	assert.NoError(t, err)

	passwords := []models.PasswordCredential{
		func() models.PasswordCredential {
			pId, _ := uuid.NewV4()
			return models.PasswordCredential{
				ID:        pId,
				UserId:    userId,
				Password:  string(hashedPassword),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}(),
	}
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := `{"user_id": "ec4ef049-5b88-4321-a173-21b0eff06a04", "password": "wrongpassword"}`

	req := httptest.NewRequest(http.MethodPost, "/password/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, nil, passwords, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	err = handler.Login(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	}
}

func TestPasswordHandler_Login_NonExistingUser(t *testing.T) {
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	body := `{"user_id": "ec4ef049-5b88-4321-a173-21b0eff06a04", "password": "wrongpassword"}`

	req := httptest.NewRequest(http.MethodPost, "/password/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister([]models.User{}, nil, nil, nil, nil, []models.PasswordCredential{}, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{}, &config.Config{}, test.NewAuditLogger())

	err := handler.Login(c)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
	}
}

// TestMaxPasswordLength bcrypt since version 0.5.0 only accepts passwords at least 72 bytes long. This test documents this behaviour.
func TestMaxPasswordLength(t *testing.T) {
	tests := []struct {
		name             string
		creationPassword string
		loginPassword    string
		wantErr          bool
	}{
		{
			name:             "login password 72 bytes long",
			creationPassword: "012345678901234567890123456789012345678901234567890123456789012345678901",
			loginPassword:    "012345678901234567890123456789012345678901234567890123456789012345678901",
			wantErr:          false,
		},
		{
			name:             "login password over 72 bytes long",
			creationPassword: "0123456789012345678901234567890123456789012345678901234567890123456789012",
			loginPassword:    "0123456789012345678901234567890123456789012345678901234567890123456789012",
			wantErr:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hash, err := bcrypt.GenerateFromPassword([]byte(test.creationPassword), 12)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = bcrypt.CompareHashAndPassword(hash, []byte(test.loginPassword))
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
