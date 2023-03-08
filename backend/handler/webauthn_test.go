package handler

import (
	"encoding/json"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var userId = "ec4ef049-5b88-4321-a173-21b0eff06a04"
var userIdBytes = []byte{0xec, 0x4e, 0xf0, 0x49, 0x5b, 0x88, 0x43, 0x21, 0xa1, 0x73, 0x21, 0xb0, 0xef, 0xf0, 0x6a, 0x4}

func TestNewWebauthnHandler(t *testing.T) {
	p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler, err := NewWebauthnHandler(&defaultConfig, p, sessionManager{}, test.NewAuditLogger())
	assert.NoError(t, err)
	assert.NotEmpty(t, handler)
}

func TestWebauthnHandler_BeginRegistration(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/webauthn/registration/initialize", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId)
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, credentials, sessionData, nil, nil, nil, nil, nil)
	handler, err := NewWebauthnHandler(&defaultConfig, p, sessionManager{}, test.NewAuditLogger())
	require.NoError(t, err)

	if assert.NoError(t, handler.BeginRegistration(c)) {
		creationOptions := protocol.CredentialCreation{}
		err = json.Unmarshal(rec.Body.Bytes(), &creationOptions)
		assert.NoError(t, err)
		assert.NotEmpty(t, creationOptions.Response.Challenge)
		assert.Equal(t, userIdBytes, []byte(creationOptions.Response.User.ID))
		assert.Equal(t, defaultConfig.Webauthn.RelyingParty.Id, creationOptions.Response.RelyingParty.ID)
		assert.Equal(t, creationOptions.Response.AuthenticatorSelection.ResidentKey, protocol.ResidentKeyRequirementRequired)
		assert.Equal(t, creationOptions.Response.AuthenticatorSelection.UserVerification, protocol.VerificationRequired)
		assert.True(t, *creationOptions.Response.AuthenticatorSelection.RequireResidentKey)
	}
}

func TestWebauthnHandler_FinishRegistration(t *testing.T) {
	body := `{
"id": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"rawId": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"type": "public-key",
"response": {
"attestationObject": "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVjeSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFYmehnq3OAAI1vMYKZIsLJfHwVQMAWgGhXZHA-Erj4xfo8FKEcB_PmR7mOUVuOn7GZhLwV-kTSh2hrVc6QE7NOikFYXiDo2M_mJ3huHJkDnnc5dHtIxfedbpMdex5fY3hoFs-fwymQjtdqdvti5c4x6UBAgMmIAEhWCDxvVrRgK4vpnr6JxTx-KfpSNyQUtvc47ryryZmj-P5kSJYIDox8N9bHQBrxN-b5kXqfmj3GwAJW7nNCh8UPbus3B6I",
"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoidE9yTkRDRDJ4UWY0ekZqRWp3eGFQOGZPRXJQM3p6MDhyTW9UbEpHdG5LVSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0"
}
}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/webauthn/registration/finalize", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token := jwt.New()
	err := token.Set(jwt.SubjectKey, userId)
	require.NoError(t, err)
	c.Set("session", token)

	p := test.NewPersister(users, nil, nil, nil, sessionData, nil, nil, nil, nil, nil)
	handler, err := NewWebauthnHandler(&defaultConfig, p, sessionManager{}, test.NewAuditLogger())
	require.NoError(t, err)

	if assert.NoError(t, handler.FinishRegistration(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assert.Regexp(t, `{"credential_id":".*"}`, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/webauthn/registration/finalize", strings.NewReader(body))
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	token2 := jwt.New()
	err = token.Set(jwt.SubjectKey, userId)
	require.NoError(t, err)
	c2.Set("session", token2)

	err = handler.FinishRegistration(c2)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
		assert.Equal(t, "Stored challenge and received challenge do not match: sessionData not found", err.Error())
	}
}

func TestWebauthnHandler_BeginAuthentication(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/webauthn/login/initialize", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, nil, sessionData, nil, nil, nil, nil, nil)
	handler, err := NewWebauthnHandler(&defaultConfig, p, sessionManager{}, test.NewAuditLogger())
	require.NoError(t, err)

	if assert.NoError(t, handler.BeginAuthentication(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assertionOptions := protocol.CredentialAssertion{}
		err = json.Unmarshal(rec.Body.Bytes(), &assertionOptions)
		assert.NoError(t, err)
		assert.NotEmpty(t, assertionOptions.Response.Challenge)
		assert.Equal(t, assertionOptions.Response.UserVerification, protocol.VerificationRequired)
		assert.Equal(t, defaultConfig.Webauthn.RelyingParty.Id, assertionOptions.Response.RelyingPartyID)
	}
}

func TestWebauthnHandler_FinishAuthentication(t *testing.T) {
	body := `{
"id": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"rawId": "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
"type": "public-key",
"response": {
"authenticatorData": "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFYmezOw",
"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiZ0tKS21oOTB2T3BZTzU1b0hwcWFIWF9vTUNxNG9UWnQtRDBiNnRlSXpyRSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
"signature": "MEYCIQDi2vYVspG6pf38I4GyQCPOojGbvX4nwSPXCi0hm80twAIhAO3EWjhAnj0UpjU_l0AH5sEh3zq4LDvkvo3AUqaqfGYD",
"userHandle": "7E7wSVuIQyGhcyGw7_BqBA"
}
}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/webauthn/login/finalize", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := test.NewPersister(users, nil, nil, credentials, sessionData, nil, nil, nil, nil, nil)
	handler, err := NewWebauthnHandler(&defaultConfig, p, sessionManager{}, test.NewAuditLogger())
	require.NoError(t, err)

	if assert.NoError(t, handler.FinishAuthentication(c)) {
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		cookies := rec.Result().Cookies()
		if assert.NotEmpty(t, cookies) {
			for _, cookie := range cookies {
				if cookie.Name == "hanko" {
					assert.Equal(t, userId, cookie.Value)
				}
			}
		}
	}

	req2 := httptest.NewRequest(http.MethodPost, "/webauthn/login/finalize", strings.NewReader(body))
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = handler.FinishAuthentication(c2)
	if assert.Error(t, err) {
		httpError := dto.ToHttpError(err)
		assert.Equal(t, http.StatusUnauthorized, httpError.Code)
		assert.Equal(t, "Stored challenge and received challenge do not match: sessionData not found", err.Error())
	}
}

var defaultConfig = config.Config{
	Webauthn: config.WebauthnSettings{
		RelyingParty: config.RelyingParty{
			Id:          "localhost",
			DisplayName: "Test Relying Party",
			Icon:        "",
			Origins:     []string{"http://localhost:8080"},
		},
		Timeout: 60000,
	},
}

type sessionManager struct {
}

func (s sessionManager) GenerateJWT(uuid uuid.UUID) (string, error) {
	return userId, nil
}

func (s sessionManager) GenerateCookie(token string) (*http.Cookie, error) {
	return &http.Cookie{
		Name:     "hanko",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func (s sessionManager) DeleteCookie() (*http.Cookie, error) {
	return &http.Cookie{
		Name:     "hanko",
		Value:    "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}, nil
}

func (s sessionManager) Verify(token string) (jwt.Token, error) {
	return nil, nil
}

var credentials = []models.WebauthnCredential{
	func() models.WebauthnCredential {
		uId, _ := uuid.FromString(userId)
		aaguid, _ := uuid.FromString("adce0002-35bc-c60a-648b-0b25f1f05503")
		return models.WebauthnCredential{
			ID:              "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjH",
			UserId:          uId,
			PublicKey:       "pQECAyYgASFYIPG9WtGAri-mevonFPH4p-lI3JBS29zjuvKvJmaP4_mRIlggOjHw31sdAGvE35vmRep-aPcbAAlbuc0KHxQ9u6zcHog",
			AttestationType: "none",
			AAGUID:          aaguid,
			SignCount:       1650958750,
			CreatedAt:       time.Time{},
			UpdatedAt:       time.Time{},
		}
	}(),
	func() models.WebauthnCredential {
		uId, _ := uuid.FromString(userId)
		aaguid, _ := uuid.FromString("adce0002-35bc-c60a-648b-0b25f1f05503")
		return models.WebauthnCredential{
			ID:              "AaFdkcD4SuPjF-jwUoRwH8-ZHuY5RW46fsZmEvBX6RNKHaGtVzpATs06KQVheIOjYz-YneG4cmQOedzl0e0jF951ukx17Hl9jeGgWz5_DKZCO12p2-2LlzjK",
			UserId:          uId,
			PublicKey:       "pQECAyYgASFYIPG9WtGAri-mevonFPH4p-lI3JBS29zjuvKvJmaP4_mRIlggOjHw31sdAGvE35vmRep-aPcbAAlbuc0KHxQ9u6zcHoj",
			AttestationType: "none",
			AAGUID:          aaguid,
			SignCount:       1650958750,
			CreatedAt:       time.Time{},
			UpdatedAt:       time.Time{},
		}
	}(),
}

var sessionData = []models.WebauthnSessionData{
	func() models.WebauthnSessionData {
		id, _ := uuid.NewV4()
		uId, _ := uuid.FromString(userId)
		return models.WebauthnSessionData{
			ID:                 id,
			Challenge:          "tOrNDCD2xQf4zFjEjwxaP8fOErP3zz08rMoTlJGtnKU",
			UserId:             uId,
			UserVerification:   string(protocol.VerificationRequired),
			CreatedAt:          time.Time{},
			UpdatedAt:          time.Time{},
			Operation:          models.WebauthnOperationRegistration,
			AllowedCredentials: nil,
		}
	}(),
	func() models.WebauthnSessionData {
		id, _ := uuid.NewV4()
		return models.WebauthnSessionData{
			ID:                 id,
			Challenge:          "gKJKmh90vOpYO55oHpqaHX_oMCq4oTZt-D0b6teIzrE",
			UserId:             uuid.UUID{},
			UserVerification:   string(protocol.VerificationRequired),
			CreatedAt:          time.Time{},
			UpdatedAt:          time.Time{},
			Operation:          models.WebauthnOperationAuthentication,
			AllowedCredentials: nil,
		}
	}(),
}

var uId, _ = uuid.FromString(userId)

var emails = []models.Email{
	{
		ID:           uId,
		Address:      "john.doe@example.com",
		PrimaryEmail: &models.PrimaryEmail{ID: uId},
	},
}

var users = []models.User{
	func() models.User {
		return models.User{
			ID:        uId,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Emails:    emails,
		}
	}(),
}
