package handler

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/test"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPasswordSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(passwordSuite))
}

type passwordSuite struct {
	test.Suite
}

func (s *passwordSuite) TestPasswordHandler_Set_Create() {
	if testing.Short() {
		s.T().Skip("skipping in short mode")
	}

	userWithNoPassword := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")
	userWithPassword := uuid.FromStringOrNil("38bf5a00-d7ea-40a5-a5de-48722c148925")
	unknownUser := uuid.FromStringOrNil("6a565180-2366-45b1-8785-39f7902c7f2e")

	cfg := &test.DefaultConfig
	cfg.Password.Enabled = true
	cfg.Password.MinPasswordLength = 8

	tests := []struct {
		name         string
		body         string
		userId       uuid.UUID
		expectedCode int
	}{
		{
			name:         "should create a password successful",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "verybadpassword"}`, userWithNoPassword),
			userId:       userWithNoPassword,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "should update a password successful",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "verybadpassword"}`, userWithPassword),
			userId:       userWithPassword,
			expectedCode: http.StatusOK,
		},
		{
			name:         "should not create a password that is too short",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "very"}`, userWithNoPassword),
			userId:       userWithNoPassword,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "should not create a password that is too long",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "thisIsAVeryLongPasswordThatIsUsedToTestIfAnErrorWillBeReturnedForTooLongPasswords"}`, userWithNoPassword),
			userId:       userWithNoPassword,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "should not create a password for an unknown user",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "verybadpassword"}`, unknownUser),
			userId:       unknownUser,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "should not create a password for a different user",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "verybadpassword"}`, userWithNoPassword),
			userId:       userWithPassword,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			s.SetupTest()
			defer s.TearDownTest()

			err := s.LoadFixtures("../test/fixtures/password")
			s.Require().NoError(err)

			sessionManager := s.GetDefaultSessionManager()
			token, err := sessionManager.GenerateJWT(currentTest.userId)
			s.Require().NoError(err)
			cookie, err := sessionManager.GenerateCookie(token)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, "/password", strings.NewReader(currentTest.body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()

			e := NewPublicRouter(cfg, s.Storage, nil, nil)
			e.ServeHTTP(rec, req)

			s.Equal(currentTest.expectedCode, rec.Code)
		})
	}
}

func (s *passwordSuite) TestPasswordHandler_Login() {
	if testing.Short() {
		s.T().Skip("skipping in short mode")
	}

	err := s.LoadFixtures("../test/fixtures/password")
	s.Require().NoError(err)

	userWithPassword := uuid.FromStringOrNil("38bf5a00-d7ea-40a5-a5de-48722c148925")
	unknownUser := uuid.FromStringOrNil("6a565180-2366-45b1-8785-39f7902c7f2e")

	cfg := func() *config.Config {
		cfg := test.DefaultConfig
		cfg.Password.Enabled = true
		return &cfg
	}

	tests := []struct {
		name                string
		body                string
		expectedCode        int
		cfg                 func() *config.Config
		shouldContainCookie bool
	}{
		{
			name:                "should login successful",
			body:                fmt.Sprintf(`{"user_id": "%s", "password": "SuperSecure"}`, userWithPassword),
			cfg:                 cfg,
			expectedCode:        http.StatusOK,
			shouldContainCookie: true,
		},
		{
			name: "should login successful with token in header",
			body: fmt.Sprintf(`{"user_id": "%s", "password": "SuperSecure"}`, userWithPassword),
			cfg: func() *config.Config {
				cfg := test.DefaultConfig
				cfg.Password.Enabled = true
				cfg.Session.EnableAuthTokenHeader = true
				return &cfg
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "should not login with wrong password",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "verybadpassword"}`, userWithPassword),
			cfg:          cfg,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "should not login with non existing user",
			body:         fmt.Sprintf(`{"user_id": "%s", "password": "verybadpassword"}`, unknownUser),
			cfg:          cfg,
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			req := httptest.NewRequest(http.MethodPost, "/password/login", strings.NewReader(currentTest.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := NewPublicRouter(currentTest.cfg(), s.Storage, nil, nil)
			e.ServeHTTP(rec, req)

			if s.Equal(currentTest.expectedCode, rec.Code) {
				if currentTest.shouldContainCookie {
					s.Empty(rec.Header().Get("X-Auth-Token"))
					cookies := rec.Result().Cookies()
					if s.NotEmpty(cookies) {
						for _, cookie := range cookies {
							if cookie.Name == "hanko" {
								s.Regexp(".*\\..*\\..*", cookie.Value) // check if cookie contains a jwt
							}
						}
					}
				} else if currentTest.cfg().Session.EnableAuthTokenHeader {
					s.Empty(rec.Result().Cookies())
					token := rec.Header().Get("X-Auth-Token")
					s.NotEmpty(token)
					s.Regexp(".*\\..*\\..*", token)
				}
			}
		})
	}
}

func (s *passwordSuite) GetDefaultSessionManager() session.Manager {
	jwkManager, err := jwk.NewDefaultManager(test.DefaultConfig.Secrets.Keys, s.Storage.GetJwkPersister())
	s.Require().NoError(err)
	sessionManager, err := session.NewManager(jwkManager, test.DefaultConfig)
	s.Require().NoError(err)

	return sessionManager
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
