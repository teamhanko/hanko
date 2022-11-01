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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewEmailHandler(t *testing.T) {
	emailHandler, err := NewEmailHandler(&config.Config{}, test.NewPersister(nil, nil, nil, nil, nil, nil, nil, nil, nil), sessionManager{}, test.NewAuditLogger())
	assert.NoError(t, err)
	assert.NotEmpty(t, emailHandler)
}

func TestEmailHandler_List(t *testing.T) {
	var emails []*dto.EmailResponse
	uId1, _ := uuid.NewV4()
	uId2, _ := uuid.NewV4()

	tests := []struct {
		name          string
		userId        uuid.UUID
		data          []models.Email
		expectedCount int
	}{
		{
			name:   "should return all user assigned email addresses",
			userId: uId1,
			data: []models.Email{
				{
					UserID:  uId1,
					Address: "john.doe+1@example.com",
				},
				{
					UserID:  uId1,
					Address: "john.doe+2@example.com",
				},
				{
					UserID:  uId2,
					Address: "john.doe+3@example.com",
				},
			},
			expectedCount: 2,
		},
		{
			name:   "should return an empty list when the user has no email addresses assigned",
			userId: uId2,
			data: []models.Email{
				{
					UserID:  uId1,
					Address: "john.doe+1@example.com",
				},
				{
					UserID:  uId1,
					Address: "john.doe+2@example.com",
				},
			},
			expectedCount: 0,
		},
	}

	for _, currentTest := range tests {
		t.Run(currentTest.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = dto.NewCustomValidator()
			req := httptest.NewRequest(http.MethodGet, "/user", nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			token := jwt.New()
			err := token.Set(jwt.SubjectKey, currentTest.userId.String())
			require.NoError(t, err)
			c.Set("session", token)
			p := test.NewPersister(nil, nil, nil, nil, nil, nil, nil, currentTest.data, nil)
			handler, err := NewEmailHandler(&config.Config{}, p, sessionManager{}, test.NewAuditLogger())
			assert.NoError(t, err)

			if assert.NoError(t, handler.List(c)) {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &emails))
				assert.Equal(t, currentTest.expectedCount, len(emails))
			}
		})
	}
}

func TestEmailHandler_Update(t *testing.T) {
	uId, _ := uuid.NewV4()
	emailId1, _ := uuid.NewV4()
	emailId2, _ := uuid.NewV4()
	testData := []models.User{
		{
			ID: uId,
			Emails: []models.Email{
				{
					ID:           emailId1,
					Address:      "john.doe@example.com",
					PrimaryEmail: nil,
				},
				{
					ID:           emailId2,
					Address:      "john.doe@example.com",
					PrimaryEmail: &models.PrimaryEmail{},
				},
			},
		},
	}

	isPrimary := true
	body := dto.EmailUpdateRequest{IsPrimary: &isPrimary}
	bodyJson, err := json.Marshal(body)
	require.NoError(t, err)
	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/emails/:id")
	c.SetParamNames("id")
	c.SetParamValues(emailId1.String())
	token := jwt.New()
	err = token.Set(jwt.SubjectKey, uId.String())
	require.NoError(t, err)
	c.Set("session", token)
	p := test.NewPersister(testData, nil, nil, nil, nil, nil, nil, nil, nil)
	handler, err := NewEmailHandler(&config.Config{}, p, sessionManager{}, test.NewAuditLogger())

	assert.NoError(t, err)
	assert.NoError(t, handler.Update(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestEmailHandler_Delete(t *testing.T) {
	uId, _ := uuid.NewV4()
	emailId1, _ := uuid.NewV4()
	emailId2, _ := uuid.NewV4()
	testData := []models.User{
		{
			ID: uId,
			Emails: []models.Email{
				{
					ID:           emailId1,
					Address:      "john.doe@example.com",
					PrimaryEmail: nil,
				},
				{
					ID:           emailId2,
					Address:      "john.doe@example.com",
					PrimaryEmail: &models.PrimaryEmail{},
				},
			},
		},
	}

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/emails/:id")
	c.SetParamNames("id")
	c.SetParamValues(emailId1.String())
	token := jwt.New()
	err := token.Set(jwt.SubjectKey, uId.String())
	require.NoError(t, err)
	c.Set("session", token)
	p := test.NewPersister(testData, nil, nil, nil, nil, nil, nil, nil, nil)
	handler, err := NewEmailHandler(&config.Config{}, p, sessionManager{}, test.NewAuditLogger())

	assert.NoError(t, err)
	assert.NoError(t, handler.Delete(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
