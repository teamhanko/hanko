package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/persistence/models"
	"github.com/teamhanko/hanko/test"
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
				Email:     "john.doe@example.com",
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

	p := test.NewPersister(users, nil, nil, nil, nil, []models.PasswordCredential{})
	handler := NewPasswordHandler(p, sessionManager{})

	if assert.NoError(t, handler.Set(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}

func TestPasswordHandler_Set_Update(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("verybadpassword"), 12)

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

	p := test.NewPersister(users, nil, nil, nil, nil, passwords)
	handler := NewPasswordHandler(p, sessionManager{})

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

	p := test.NewPersister([]models.User{}, nil, nil, nil, nil, []models.PasswordCredential{})
	handler := NewPasswordHandler(p, sessionManager{})

	if assert.NoError(t, handler.Set(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestPasswordHandler_Set_TokenHasWrongSubject(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
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

	p := test.NewPersister(users, nil, nil, nil, nil, passwords)
	handler := NewPasswordHandler(p, sessionManager{})

	if assert.NoError(t, handler.Set(c)) {
		assert.Equal(t, http.StatusForbidden, rec.Code)
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

	p := test.NewPersister(nil, nil, nil, nil, nil, nil)
	handler := NewPasswordHandler(p, sessionManager{})

	if assert.NoError(t, handler.Set(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		apiError := dto.ApiError{}
		err := json.Unmarshal(rec.Body.Bytes(), &apiError)
		require.NoError(t, err)
		assert.Equal(t, 2, len(apiError.ValidationErrors))
	}
}

func TestPasswordHandler_Login(t *testing.T) {
	userId, _ := uuid.FromString("ec4ef049-5b88-4321-a173-21b0eff06a04")
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

	p := test.NewPersister(users, nil, nil, nil, nil, passwords)
	handler := NewPasswordHandler(p, sessionManager{})

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
				Email:     "john.doe@example.com",
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

	p := test.NewPersister(users, nil, nil, nil, nil, passwords)
	handler := NewPasswordHandler(p, sessionManager{})

	if assert.NoError(t, handler.Login(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
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

	p := test.NewPersister([]models.User{}, nil, nil, nil, nil, []models.PasswordCredential{})
	handler := NewPasswordHandler(p, sessionManager{})

	if assert.NoError(t, handler.Login(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}
