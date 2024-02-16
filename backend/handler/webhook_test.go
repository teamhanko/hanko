package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookHandlerSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(webhookSuite))
}

type webhookSuite struct {
	test.Suite
}

func (s *webhookSuite) TestWebhookHandler_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/webhooks")
	s.Require().NoError(err)

	cfg := test.DefaultConfig

	cfg.Webhooks = config.WebhookSettings{
		Enabled: true,
		Hooks: config.Webhooks{
			config.Webhook{
				Callback: "http://lorem",
				Events:   events.Events{events.UserDelete},
			},
			config.Webhook{
				Callback: "http://ipsum",
				Events:   events.Events{events.UserCreate},
			},
		},
	}

	e := NewAdminRouter(&cfg, s.Storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var dto admin.WebhookListResponseDto
	err = json.Unmarshal(rec.Body.Bytes(), &dto)

	s.Require().NoError(err)
	s.Equal(5, len(dto.Database))
	s.Equal(2, len(dto.Config))
}

func (s *webhookSuite) TestWebhookHandler_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/webhooks")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	testBody := admin.CreateWebhookRequestDto{
		Callback: "http://lorem",
		Events: events.Events{
			events.UserDelete,
		},
	}
	testBodyJson, err := json.Marshal(testBody)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(testBodyJson))
	req.Header.Add("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	result := models.Webhook{}
	err = json.Unmarshal(rec.Body.Bytes(), &result)

	s.Require().NoError(err)
	s.Equal(testBody.Callback, result.Callback)
	s.Equal(string(testBody.Events[0]), result.WebhookEvents[0].Event)
	s.Equal(1, len(result.WebhookEvents))
	s.Require().NotNil(result.ID)
	s.Require().NotNil(result.WebhookEvents[0].ID)

	err = e.Close()
	s.Require().NoError(err)
}

func (s *webhookSuite) TestWebhookHandler_CreateWithParams() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name           string
		callback       string
		events         events.Events
		expectedStatus int
	}{
		{
			name:           "success",
			callback:       "http://lorem.ipsum",
			events:         events.Events{events.UserDelete},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty callback",
			callback:       "",
			events:         events.Events{events.UserDelete},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing callback",
			events:         events.Events{events.UserDelete},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "wrong callback",
			callback:       "lorem",
			events:         events.Events{events.UserDelete},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty events",
			callback:       "http://lorem.ipsum",
			events:         events.Events{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "wrong event",
			callback:       "http://lorem.ipsum",
			events:         events.Events{events.Event("cat")},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing event array",
			callback:       "http://lorem.ipsum",
			events:         nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())

			err := s.LoadFixtures("../test/fixtures/webhooks")
			s.Require().NoError(err)

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			testBody := admin.CreateWebhookRequestDto{
				Callback: currentTest.callback,
				Events:   currentTest.events,
			}
			testBodyJson, err := json.Marshal(testBody)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(testBodyJson))
			req.Header.Add("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatus, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))
		})
	}
}

func (s *webhookSuite) TestWebhookHandler_Delete() {
	s.Require().NoError(s.Storage.MigrateUp())

	err := s.LoadFixtures("../test/fixtures/webhooks")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	testId := "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4"
	testUuid, err := uuid.FromString(testId)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/webhooks/%s", testId), nil)

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusNoContent, rec.Code)

	persister := s.Storage.GetWebhookPersister(nil)

	entry, err := persister.Get(testUuid)
	s.Require().NoError(err)

	list, err := persister.List(true)
	s.Require().NoError(err)

	s.Require().Nil(entry)
	s.Equal(4, len(list))

	err = e.Close()
	s.Require().NoError(err)

	s.Require().NoError(s.Storage.MigrateDown(-1))
}

func (s *webhookSuite) TestWebhookHandler_DeleteWithParams() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name           string
		testId         string
		expectedStatus int
	}{
		{
			name:           "success",
			testId:         "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "empty id",
			testId:         "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non uuid",
			testId:         "lorem",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing id",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())

			err := s.LoadFixtures("../test/fixtures/webhooks")
			s.Require().NoError(err)

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/webhooks/%s", currentTest.testId), nil)

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatus, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))

		})
	}
}
func (s *webhookSuite) TestWebhookHandler_Get() {
	s.Require().NoError(s.Storage.MigrateUp())

	err := s.LoadFixtures("../test/fixtures/webhooks")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	testId := "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4"

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/webhooks/%s", testId), nil)

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var dto models.Webhook
	err = json.Unmarshal(rec.Body.Bytes(), &dto)

	s.Require().NotNil(dto)
	s.Equal(testId, dto.ID.String())
	s.Equal("http://localhost", dto.Callback)
	s.Equal(false, dto.Enabled)
	s.Equal(3, dto.Failures)

	err = e.Close()
	s.Require().NoError(err)

	s.Require().NoError(s.Storage.MigrateDown(-1))
}

func (s *webhookSuite) TestWebhookHandler_GetWithParams() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name           string
		testId         string
		expectedStatus int
	}{
		{
			name:           "success",
			testId:         "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong ID",
			testId:         "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da7",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non UUID ID",
			testId:         "lorem",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty id",
			testId:         "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Non uuid",
			testId:         "lorem",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing id",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())

			err := s.LoadFixtures("../test/fixtures/webhooks")
			s.Require().NoError(err)

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/webhooks/%s", currentTest.testId), nil)

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatus, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))

		})
	}
}
func (s *webhookSuite) TestWebhookHandler_Update() {
	s.Require().NoError(s.Storage.MigrateUp())

	err := s.LoadFixtures("../test/fixtures/webhooks")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	testId := "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4"
	testUuid, err := uuid.FromString(testId)
	s.Require().NoError(err)

	updateDto := admin.UpdateWebhookRequestDto{
		GetWebhookRequestDto: admin.GetWebhookRequestDto{
			ID: testId,
		},
		CreateWebhookRequestDto: admin.CreateWebhookRequestDto{
			Callback: "https://ipsum.magna/lorem",
			Events: events.Events{
				events.UserDelete,
			},
		},
		Enabled: true,
	}
	updateJson, err := json.Marshal(updateDto)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/webhooks/%s", testId), bytes.NewReader(updateJson))
	req.Header.Add("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var result models.Webhook
	err = json.Unmarshal(rec.Body.Bytes(), &result)

	// Result Check
	s.Require().NotNil(result)
	s.Equal(testId, result.ID.String())
	s.Equal(updateDto.Callback, result.Callback)
	s.Equal(updateDto.Enabled, result.Enabled)
	s.Equal(0, result.Failures)
	s.Require().True(result.ExpiresAt.After(time.Now().Add(29 * 24 * time.Hour))) // 30 Days
	s.Require().True(result.CreatedAt.Before(result.UpdatedAt))

	dbHook, err := s.Storage.GetWebhookPersister(nil).Get(testUuid)
	s.Require().NoError(err)
	s.Equal(updateDto.Callback, dbHook.Callback)
	s.Equal(updateDto.Enabled, dbHook.Enabled)
	s.Equal(0, dbHook.Failures)
	s.Require().True(dbHook.ExpiresAt.After(time.Now().Add(29 * 24 * time.Hour))) // 30 Days
	s.Require().True(dbHook.CreatedAt.Before(dbHook.UpdatedAt))

	err = e.Close()
	s.Require().NoError(err)

	s.Require().NoError(s.Storage.MigrateDown(-1))
}

func (s *webhookSuite) TestWebhookHandler_UpdateWithParams() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tests := []struct {
		name           string
		testId         string
		callback       string
		events         events.Events
		enabled        bool
		expectedStatus int
	}{
		{
			name:     "success",
			testId:   "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			callback: "https://lorem.ipsum.et",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:     "wrong ID",
			testId:   "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da7",
			callback: "https://lorem.ipsum.et",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "empty ID",
			testId:   "",
			callback: "https://lorem.ipsum.et",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "non UUID ID",
			testId:   "lorem",
			callback: "https://lorem.ipsum.et",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "missing ID",
			callback: "https://lorem.ipsum.et",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "empty Callback",
			testId:   "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			callback: "",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "wrong Callback",
			testId:   "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			callback: "lorem",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "missing Callback",
			testId: "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			events: events.Events{
				events.UserDelete,
			},
			enabled:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty events",
			testId:         "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			callback:       "https://lorem.ipsum.et",
			events:         events.Events{},
			enabled:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing events",
			testId:         "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			callback:       "https://lorem.ipsum.et",
			enabled:        true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "missing enable",
			testId:   "8b00da9a-cacf-45ea-b25d-c1ce0f0d7da4",
			callback: "https://lorem.ipsum.et",
			events: events.Events{
				events.UserDelete,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.Require().NoError(s.Storage.MigrateUp())

			err := s.LoadFixtures("../test/fixtures/webhooks")
			s.Require().NoError(err)

			e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

			updateDto := admin.UpdateWebhookRequestDto{
				GetWebhookRequestDto: admin.GetWebhookRequestDto{
					ID: currentTest.testId,
				},
				CreateWebhookRequestDto: admin.CreateWebhookRequestDto{
					Callback: currentTest.callback,
					Events:   currentTest.events,
				},
				Enabled: currentTest.enabled,
			}
			updateJson, err := json.Marshal(updateDto)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/webhooks/%s", currentTest.testId), bytes.NewReader(updateJson))
			req.Header.Add("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatus, rec.Code)

			err = e.Close()
			s.Require().NoError(err)

			s.Require().NoError(s.Storage.MigrateDown(-1))

		})
	}
}
