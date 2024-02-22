package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmailSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(emailSuite))
}

type emailSuite struct {
	test.Suite
}

func (s *emailSuite) TestEmailHandler_New() {
	emailHandler := NewEmailHandler(&config.Config{}, s.Storage, sessionManager{}, test.NewAuditLogger())
	s.NotEmpty(emailHandler)
}

func (s *emailSuite) TestEmailHandler_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil, nil)

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	s.Require().NoError(err)
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig)
	s.Require().NoError(err)

	tests := []struct {
		name          string
		userId        uuid.UUID
		expectedCount int
	}{
		{
			name:          "should return all user assigned email addresses",
			userId:        uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"),
			expectedCount: 3,
		},
		{
			name:          "should return an empty list when the user has no email addresses assigned",
			userId:        uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
			expectedCount: 0,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			token, err := sessionManager.GenerateJWT(currentTest.userId)
			s.Require().NoError(err)
			cookie, err := sessionManager.GenerateCookie(token)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodGet, "/emails", nil)
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if s.Equal(http.StatusOK, rec.Code) {
				var emails []*dto.EmailResponse
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &emails))
				s.Equal(currentTest.expectedCount, len(emails))
			}
		})
	}
}

func (s *emailSuite) TestEmailHandler_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	auditLogRecordsKey := "email_created"

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	tests := []struct {
		name                 string
		email                string
		userId               uuid.UUID
		maxNumberOfAddresses int
		requiresVerification bool
		expectedStatusCode   int
		upsertsRecords       bool
		expectedEmailUserId  uuid.UUID
	}{
		{
			name:                 "should reject the request when the user has already reached the maximum number of email addresses",
			email:                "rejection1@example.com",
			userId:               uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"),
			maxNumberOfAddresses: 0,
			requiresVerification: false,
			expectedStatusCode:   409,
			upsertsRecords:       false,
		},
		{
			name:                 "should error if email address is already in use",
			email:                "john.doe@example.com",
			userId:               uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
			requiresVerification: false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   400,
			upsertsRecords:       false,
			expectedEmailUserId:  uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"),
		},
		{
			name:                 "should assign the email address to the user if not yet assigned and does not require verification",
			email:                "john.doe+6@example.com",
			userId:               uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
			requiresVerification: false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   200,
			upsertsRecords:       true,
			expectedEmailUserId:  uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
		},
		{
			name:                 "should not assign the email address to the user if not yet assigned and requires verification",
			email:                "john.doe+7@example.com",
			userId:               uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
			requiresVerification: true,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   200,
			upsertsRecords:       false,
			expectedEmailUserId:  uuid.Nil,
		},
		{
			name:                 "should create email record with nil user ID if verification required",
			email:                "test.email.verification@example.com",
			userId:               uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
			requiresVerification: true,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   200,
			upsertsRecords:       true,
			expectedEmailUserId:  uuid.Nil,
		},
		{
			name:                 "should create email record with user ID if verification not required",
			email:                "test.email.noverification@example.com",
			userId:               uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
			requiresVerification: false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   200,
			upsertsRecords:       true,
			expectedEmailUserId:  uuid.FromStringOrNil("d41df4b7-c055-45e6-9faf-61aa92a4032e"),
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := test.DefaultConfig
			cfg.AuditLog.Storage.Enabled = true
			cfg.Emails.RequireVerification = currentTest.requiresVerification
			cfg.Emails.MaxNumOfAddresses = currentTest.maxNumberOfAddresses
			e := NewPublicRouter(&cfg, s.Storage, nil, nil)
			jwkManager, err := jwk.NewDefaultManager(cfg.Secrets.Keys, s.Storage.GetJwkPersister())
			s.Require().NoError(err)
			sessionManager, err := session.NewManager(jwkManager, cfg)
			s.Require().NoError(err)

			token, err := sessionManager.GenerateJWT(currentTest.userId)
			s.Require().NoError(err)
			cookie, err := sessionManager.GenerateCookie(token)
			s.Require().NoError(err)

			body := dto.EmailCreateRequest{
				Address: currentTest.email,
			}
			bodyJson, err := json.Marshal(body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/emails", bytes.NewReader(bodyJson))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()

			auditLogsCountBefore := s.getAuditLogRecordsCount(auditLogRecordsKey)

			e.ServeHTTP(rec, req)

			auditLogsCountAfter := s.getAuditLogRecordsCount(auditLogRecordsKey)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			email, err := s.Storage.GetEmailPersister().FindByAddress(currentTest.email)
			s.Require().NoError(err)

			if currentTest.upsertsRecords {
				s.NotNil(email)
			}

			if email != nil {
				if currentTest.expectedEmailUserId != uuid.Nil {
					s.Equal(currentTest.expectedEmailUserId, *email.UserID)
				} else {
					s.Nil(email.UserID)
				}
			}

			if rec.Code == http.StatusOK {
				s.Equal(auditLogsCountBefore+1, auditLogsCountAfter)
			} else {
				s.Equal(auditLogsCountBefore, auditLogsCountAfter)
			}
		})
	}
}

func (s *emailSuite) TestEmailHandler_SetPrimaryEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil, nil)

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	s.Require().NoError(err)
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig)
	s.Require().NoError(err)

	oldPrimaryEmailId := uuid.FromStringOrNil("51b7c175-ceb6-45ba-aae6-0092221c1b84")
	newPrimaryEmailId := uuid.FromStringOrNil("8bb4c8a7-a3e6-48bb-b54f-20e3b485ab33")
	userId := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")

	token, err := sessionManager.GenerateJWT(userId)
	s.NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.NoError(err)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/emails/%s/set_primary", newPrimaryEmailId), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	if s.Equal(http.StatusNoContent, rec.Code) {
		emails, err := s.Storage.GetEmailPersister().FindByUserId(userId)
		s.Require().NoError(err)

		s.Equal(3, len(emails))
		for _, email := range emails {
			if email.ID == newPrimaryEmailId {
				s.True(email.IsPrimary())
			} else if email.ID == oldPrimaryEmailId {
				s.False(email.IsPrimary())
			}
		}
	}
}

func (s *emailSuite) TestEmailHandler_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	e := NewPublicRouter(&test.DefaultConfig, s.Storage, nil, nil)
	userId := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")

	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	s.Require().NoError(err)
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig)
	s.Require().NoError(err)

	token, err := sessionManager.GenerateJWT(userId)
	s.NoError(err)
	cookie, err := sessionManager.GenerateCookie(token)
	s.NoError(err)

	tests := []struct {
		name          string
		emailId       uuid.UUID
		responseCode  int
		expectedCount int
	}{
		{
			name:          "should delete the email address",
			emailId:       uuid.FromStringOrNil("8bb4c8a7-a3e6-48bb-b54f-20e3b485ab33"),
			responseCode:  http.StatusNoContent,
			expectedCount: 2,
		},
		{
			name:          "should not delete the primary email address",
			emailId:       uuid.FromStringOrNil("51b7c175-ceb6-45ba-aae6-0092221c1b84"),
			responseCode:  http.StatusConflict,
			expectedCount: 2,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/emails/%s", currentTest.emailId), nil)
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)
			if s.Equal(currentTest.responseCode, rec.Code) {
				emails, err := s.Storage.GetEmailPersister().FindByUserId(userId)
				s.Require().NoError(err)
				s.Equal(currentTest.expectedCount, len(emails))
			}
		})
	}

}

func (s *emailSuite) getAuditLogRecordsCount(code string) int {
	logs, lerr := s.Storage.GetAuditLogPersister().List(0, 0, nil, nil, []string{code}, "", "", "", "")
	s.Require().NoError(lerr)
	return len(logs)
}
