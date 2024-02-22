package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmailAdminSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(emailAdminSuite))
}

type emailAdminSuite struct {
	test.Suite
}

func (s *emailAdminSuite) TestEmailAdminHandler_New() {
	emailHandler := NewEmailAdminHandler(&config.Config{}, s.Storage)
	s.NotEmpty(emailHandler)
}

func (s *emailAdminSuite) TestEmailAdminHandler_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	e := NewAdminRouter(&test.DefaultConfig, s.Storage, nil)

	tests := []struct {
		name               string
		userId             string
		expectedCount      int
		expectedStatusCode int
	}{
		{
			name:          "should return all user assigned email addresses",
			userId:        "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedCount: 3,
		},
		{
			name:          "should return an empty list when the user has no email addresses assigned",
			userId:        "d41df4b7-c055-45e6-9faf-61aa92a4032e",
			expectedCount: 0,
		},
		{
			name:               "should fail on non uuid",
			userId:             "d41df4b7-c055-45e6-9faf-61aa92a4032",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail on empty",
			userId:             "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to find non existing user",
			userId:             "d41df4b7-c055-45e6-9faf-61aa92a4032f",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			testDto := admin.ListEmailRequestDto{
				UserId: currentTest.userId,
			}
			testJson, err := json.Marshal(testDto)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/emails", currentTest.userId), bytes.NewReader(testJson))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if http.StatusOK == rec.Code {
				var emails []*dto.EmailResponse
				s.NoError(json.Unmarshal(rec.Body.Bytes(), &emails))
				s.Equal(currentTest.expectedCount, len(emails))
			} else {
				s.Require().Equal(currentTest.expectedStatusCode, rec.Code)
			}
		})
	}
}

func (s *emailAdminSuite) TestEmailAdminHandler_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	tests := []struct {
		name                 string
		email                string
		userId               string
		maxNumberOfAddresses int
		isVerified           bool
		expectedStatusCode   int
	}{
		{
			name:                 "should reject the request when the user has already reached the maximum number of email addresses",
			email:                "rejection1@example.com",
			userId:               "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			maxNumberOfAddresses: 0,
			isVerified:           false,
			expectedStatusCode:   http.StatusConflict,
		},
		{
			name:                 "should error if email address is already in use",
			email:                "john.doe@example.com",
			userId:               "d41df4b7-c055-45e6-9faf-61aa92a4032e",
			isVerified:           false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "should assign the email address to the user if not yet assigned and does not require verification",
			email:                "john.doe+6@example.com",
			userId:               "d41df4b7-c055-45e6-9faf-61aa92a4032e",
			isVerified:           false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusCreated,
		},
		{
			name:                 "should create email record with nil user ID if verification required",
			email:                "test.email.verification@example.com",
			userId:               "d41df4b7-c055-45e6-9faf-61aa92a4032e",
			isVerified:           true,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusCreated,
		},
		{
			name:                 "should create email record with user ID if verification not required",
			email:                "test.email.noverification@example.com",
			userId:               "d41df4b7-c055-45e6-9faf-61aa92a4032e",
			isVerified:           false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusCreated,
		},
		{
			name:                 "should fail to create email record with missing user id",
			email:                "test.email.noverification@example.com",
			userId:               "",
			isVerified:           false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "should fail to create email record with non uuid user id",
			email:                "test.email.noverification@example.com",
			userId:               "lorem",
			isVerified:           false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "should fail to create email record with wrong user id",
			email:                "test.email.noverification@example.com",
			userId:               "d41df4b7-c055-45e6-9faf-61aa92a4032f",
			isVerified:           false,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusNotFound,
		},
		{
			name:                 "should create verified email record with wrong user id",
			email:                "test.email.noverification@example.com",
			userId:               "d41df4b7-c055-45e6-9faf-61aa92a4032e",
			isVerified:           true,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "should fail to create email with missing email",
			email:                "",
			userId:               "d41df4b7-c055-45e6-9faf-61aa92a4032e",
			isVerified:           true,
			maxNumberOfAddresses: 100,
			expectedStatusCode:   http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := test.DefaultConfig
			cfg.Emails.MaxNumOfAddresses = currentTest.maxNumberOfAddresses

			e := NewAdminRouter(&cfg, s.Storage, nil)

			body := admin.CreateEmailRequestDto{
				ListEmailRequestDto: admin.ListEmailRequestDto{
					UserId: currentTest.userId,
				},
				CreateEmail: admin.CreateEmail{
					Address:    currentTest.email,
					IsVerified: currentTest.isVerified,
				},
			}
			bodyJson, err := json.Marshal(body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/emails", currentTest.userId), bytes.NewReader(bodyJson))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			if rec.Code == http.StatusOK {
				email, err := s.Storage.GetEmailPersister().FindByAddress(currentTest.email)
				s.Require().NoError(err)

				if email != nil {
					s.Equal(currentTest.userId, email.UserID.String())
					s.Equal(currentTest.isVerified, email.Verified)
				}
			}

		})
	}
}

func (s *emailAdminSuite) TestEmailAdminHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	tests := []struct {
		name               string
		emailId            string
		userId             string
		expectedStatusCode int
	}{
		{
			name:               "should get a single email",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should fail to get an email for a non existent user",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a6",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to get an email for a wrong user UUID",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "lorem",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to get an email for an empty user UUID",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to get a non existent email",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1ff",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to get an empty email uuid",
			emailId:            "",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to get a wrong email uuid",
			emailId:            "lorem",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := test.DefaultConfig

			e := NewAdminRouter(&cfg, s.Storage, nil)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/emails/%s", currentTest.userId, currentTest.emailId), nil)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			if rec.Code == http.StatusOK {
				var email admin.Email
				err := json.Unmarshal(rec.Body.Bytes(), &email)
				s.Require().NoError(err)

				s.Require().NotNil(email)
				s.Require().Equal(currentTest.emailId, email.ID.String())
			}
		})
	}
}

func (s *emailAdminSuite) TestEmailAdminHandler_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	tests := []struct {
		name               string
		emailId            string
		userId             string
		expectedStatusCode int
	}{
		{
			name:               "should delete an email",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should fail to delete an email for a non existent user",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a6",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to delete an email for a wrong user UUID",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "lorem",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to delete an email for an empty user UUID",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to delete a non existent email",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1ff",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to delete an empty email uuid",
			emailId:            "",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to delete a wrong email uuid",
			emailId:            "lorem",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to delete a primary email",
			emailId:            "51b7c175-ceb6-45ba-aae6-0092221c1b84",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := test.DefaultConfig

			e := NewAdminRouter(&cfg, s.Storage, nil)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/emails/%s", currentTest.userId, currentTest.emailId), nil)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			if rec.Code == http.StatusOK {
				var email admin.Email
				err := json.Unmarshal(rec.Body.Bytes(), &email)
				s.Require().NoError(err)

				s.Require().NotNil(email)
				s.Require().Equal(currentTest.emailId, email.ID.String())
			}
		})
	}
}

func (s *emailAdminSuite) TestEmailAdminHandler_SetPrimaryEmail() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/email")
	s.Require().NoError(err)

	tests := []struct {
		name               string
		emailId            string
		userId             string
		isAlreadyPrimary   bool
		expectedStatusCode int
	}{
		{
			name:               "should set an email as primary",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			isAlreadyPrimary:   false,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "should fail to set an email as primary for a non existent user",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a6",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to set an email as primary for a wrong user UUID",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "lorem",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to set an email as primary for an empty user UUID",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1fe",
			userId:             "",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to set an email as primary for a non existent email",
			emailId:            "f194ee0f-dd1a-48f7-8766-c67e4d6cd1ff",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "should fail to set an email as primary for an empty email uuid",
			emailId:            "",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should fail to set an email as primary for a wrong email uuid",
			emailId:            "lorem",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Should not change a primary email",
			emailId:            "51b7c175-ceb6-45ba-aae6-0092221c1b84",
			userId:             "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			expectedStatusCode: http.StatusNoContent,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := test.DefaultConfig

			e := NewAdminRouter(&cfg, s.Storage, nil)

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/emails/%s/set_primary", currentTest.userId, currentTest.emailId), nil)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedStatusCode, rec.Code)

			if rec.Code == http.StatusNoContent {
				userUuid, err := uuid.FromString(currentTest.userId)
				s.Require().NoError(err)

				emails, err := s.Storage.GetEmailPersister().FindByUserId(userUuid)
				s.Require().NoError(err)

				s.Equal(3, len(emails))
				for _, email := range emails {
					if email.ID.String() == currentTest.emailId {
						s.True(email.IsPrimary())
					}
				}
			}
		})
	}
}
