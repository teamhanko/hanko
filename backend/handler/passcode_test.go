package handler

import (
	"bytes"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPasscodeSuite(t *testing.T) {
	s := new(passcodeSuite)
	s.WithEmailServer = true
	suite.Run(t, s)
}

type passcodeSuite struct {
	test.Suite
}

func (s *passcodeSuite) TestPasscodeHandler_Init() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}
	err := s.LoadFixtures("../test/fixtures/passcode")
	s.Require().NoError(err)

	cfg := func() *config.Config {
		cfg := &test.DefaultConfig
		cfg.Passcode.Smtp.Host = "localhost"
		cfg.Passcode.Smtp.Port = s.EmailServer.SmtpPort
		return cfg
	}

	e := NewPublicRouter(cfg(), s.Storage, nil)

	emailId := "51b7c175-ceb6-45ba-aae6-0092221c1b84"
	unknownEmailId := "83618f24-2db8-4ea2-b370-ac8335f782d8"
	tests := []struct {
		name                 string
		body                 dto.PasscodeInitRequest
		expectedStatusCode   int
		expectedEmailAddress string
	}{
		{
			name: "with userID and emailID",
			body: dto.PasscodeInitRequest{
				UserId:  "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
				EmailId: &emailId,
			},
			expectedStatusCode:   http.StatusOK,
			expectedEmailAddress: "john.doe@example.com",
		},
		{
			name: "only with userID",
			body: dto.PasscodeInitRequest{
				UserId: "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
			},
			expectedStatusCode:   http.StatusOK,
			expectedEmailAddress: "john.doe@example.com",
		},
		{
			name: "with unknown userID",
			body: dto.PasscodeInitRequest{
				UserId: "83618f24-2db8-4ea2-b370-ac8335f782d8",
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "with unknown emailID",
			body: dto.PasscodeInitRequest{
				UserId:  "b5dd5267-b462-48be-b70d-bcd6f1bbe7a5",
				EmailId: &unknownEmailId,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			bodyJson, err := json.Marshal(currentTest.body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/passcode/login/initialize", bytes.NewReader(bodyJson))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if s.Equal(currentTest.expectedStatusCode, rec.Code) && currentTest.expectedStatusCode >= 200 && currentTest.expectedStatusCode <= 299 {
				emails, err := test.GetEmails(s.EmailServer)
				s.Require().NoError(err)
				messages := emails.MailItems

				s.Require().Greater(len(messages), 0)

				emailAddress := messages[len(messages)-1].ToAddresses[0]
				s.Equal(currentTest.expectedEmailAddress, emailAddress)
			}
		})
	}
}

func (s *passcodeSuite) TestPasscodeHandler_Finish() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}
	err := s.LoadFixtures("../test/fixtures/passcode")
	s.Require().NoError(err)

	now := time.Now().UTC()

	hashedPasscode, err := bcrypt.GenerateFromPassword([]byte("123456"), 12)

	passcode := models.Passcode{
		ID:        uuid.FromStringOrNil("a2383922-dea3-46c8-be17-85b267c0d135"),
		UserId:    uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"),
		EmailID:   uuid.FromStringOrNil("51b7c175-ceb6-45ba-aae6-0092221c1b84"),
		Ttl:       300,
		Code:      string(hashedPasscode),
		TryCount:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	passcodeWithExpiredTimeout := models.Passcode{
		ID:        uuid.FromStringOrNil("a2383922-dea3-46c8-be17-85b267c0d135"),
		UserId:    uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5"),
		EmailID:   uuid.FromStringOrNil("51b7c175-ceb6-45ba-aae6-0092221c1b84"),
		Ttl:       300,
		Code:      string(hashedPasscode),
		TryCount:  0,
		CreatedAt: now.Add(-500 * time.Second),
		UpdatedAt: now,
	}

	cfg := func() *config.Config {
		return &test.DefaultConfig
	}

	tests := []struct {
		name               string
		passcodeId         string
		retryCount         int
		passcode           models.Passcode
		code               string
		expectedStatusCode int
		cfg                func() *config.Config
	}{
		{
			name:               "finish successful",
			passcodeId:         "a2383922-dea3-46c8-be17-85b267c0d135",
			passcode:           passcode,
			code:               "123456",
			expectedStatusCode: http.StatusOK,
			cfg:                cfg,
		},
		{
			name:               "finish successful with token in header",
			passcodeId:         "a2383922-dea3-46c8-be17-85b267c0d135",
			passcode:           passcode,
			code:               "123456",
			expectedStatusCode: http.StatusOK,
			cfg: func() *config.Config {
				c := test.DefaultConfig
				c.Session.EnableAuthTokenHeader = true
				return &c
			},
		},
		{
			name:               "with wrong code",
			passcodeId:         "a2383922-dea3-46c8-be17-85b267c0d135",
			passcode:           passcode,
			code:               "654321",
			expectedStatusCode: http.StatusUnauthorized,
			cfg:                cfg,
		},
		{
			name:               "with wrong code 3 times",
			passcodeId:         "a2383922-dea3-46c8-be17-85b267c0d135",
			retryCount:         2,
			passcode:           passcode,
			code:               "654321",
			expectedStatusCode: http.StatusGone,
			cfg:                cfg,
		},
		{
			name:               "with wrong passcode ID",
			passcodeId:         "297cfc1b-98cc-4ae1-bc83-bcafc7f0e876",
			passcode:           passcode,
			code:               "123456",
			expectedStatusCode: http.StatusUnauthorized,
			cfg:                cfg,
		},
		{
			name:               "after passcode expired",
			passcodeId:         "a2383922-dea3-46c8-be17-85b267c0d135",
			passcode:           passcodeWithExpiredTimeout,
			code:               "123456",
			expectedStatusCode: http.StatusRequestTimeout,
			cfg:                cfg,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			e := NewPublicRouter(currentTest.cfg(), s.Storage, nil)

			// Setup passcode
			err := s.Storage.GetPasscodePersister().Create(currentTest.passcode)
			s.Require().NoError(err)

			body := dto.PasscodeFinishRequest{
				Id:   currentTest.passcodeId,
				Code: currentTest.code,
			}
			bodyJson, err := json.Marshal(body)
			s.Require().NoError(err)

			responseCode := 0
			var response *http.Response
			var headers http.Header
			for i := 0; i <= currentTest.retryCount; i++ {
				req := httptest.NewRequest(http.MethodPost, "/passcode/login/finalize", bytes.NewReader(bodyJson))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()

				e.ServeHTTP(rec, req)
				responseCode = rec.Code
				response = rec.Result()
				headers = rec.Header()
			}

			s.Equal(currentTest.expectedStatusCode, responseCode)

			if currentTest.cfg().Session.EnableAuthTokenHeader {
				s.Empty(response.Cookies())
				token := headers.Get("X-Auth-Token")
				s.NotEmpty(token)
				s.Regexp(".*\\..*\\..*", token)
			}

			// remove passcode
			_ = s.Storage.GetPasscodePersister().Delete(currentTest.passcode)
		})
	}
}
