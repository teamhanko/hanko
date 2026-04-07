package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/session"
	"github.com/teamhanko/hanko/backend/v2/test"
)

func TestSessionSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(sessionSuite))
}

type sessionSuite struct {
	test.Suite
}

func (s *sessionSuite) TestSessionHandler_ValidateSession_IdleExpiresAt() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	testUserID := uuid.FromStringOrNil("ec4ef049-5b88-4321-a173-21b0eff06a04")

	tests := []struct {
		name                 string
		idleTimeout          string
		sessionExpiresAt     *time.Time
		expectIdleExpiresAt  bool
		expectCappedToJWTExp bool
	}{
		{
			name:                "should return idle_expires_at when idle timeout is configured",
			idleTimeout:         "1h",
			expectIdleExpiresAt: true,
		},
		{
			name:                "should not return idle_expires_at when idle timeout is not configured",
			idleTimeout:         "0s",
			expectIdleExpiresAt: false,
		},
		{
			name:                 "should cap idle_expires_at to JWT expiration when idle timeout exceeds it",
			idleTimeout:          "24h",
			sessionExpiresAt:     timePtr(time.Now().Add(2 * time.Hour)),
			expectIdleExpiresAt:  true,
			expectCappedToJWTExp: true,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := s.setupConfig(currentTest.idleTimeout)

			// Create sess with cookie
			cookie, _ := s.createSessionWithCookie(testUserID, cfg, currentTest.sessionExpiresAt)

			e := NewPublicRouter(cfg, s.Storage, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/sessions/validate", nil)
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(http.StatusOK, rec.Code)

			var response dto.ValidateSessionResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			s.Require().NoError(err)

			s.True(response.IsValid)
			s.NotNil(response.Claims)

			if currentTest.expectIdleExpiresAt {
				s.NotNil(response.IdleExpiresAt, "idle_expires_at should be set")

				if currentTest.expectCappedToJWTExp {
					// Should be capped to JWT expiration
					s.Equal(response.Claims.Expiration.Unix(), response.IdleExpiresAt.Unix(),
						"idle_expires_at should be capped to JWT expiration")
				} else {
					// Should be approximately now + idle timeout
					expectedTime := time.Now().Add(1 * time.Hour)
					s.InDelta(expectedTime.Unix(), response.IdleExpiresAt.Unix(), 5,
						"idle_expires_at should be now + idle timeout")
				}
			} else {
				s.Nil(response.IdleExpiresAt, "idle_expires_at should not be set when idle timeout is 0")
			}
		})
	}
}

func (s *sessionSuite) TestSessionHandler_ValidateSessionFromBody_IdleExpiresAt() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	testUserID := uuid.FromStringOrNil("ec4ef049-5b88-4321-a173-21b0eff06a04")

	tests := []struct {
		name                 string
		idleTimeout          string
		sessionExpiresAt     *time.Time
		expectIdleExpiresAt  bool
		expectCappedToJWTExp bool
	}{
		{
			name:                "should return idle_expires_at when idle timeout is configured",
			idleTimeout:         "1h",
			expectIdleExpiresAt: true,
		},
		{
			name:                "should not return idle_expires_at when idle timeout is not configured",
			idleTimeout:         "0s",
			expectIdleExpiresAt: false,
		},
		{
			name:                 "should cap idle_expires_at to JWT expiration when idle timeout exceeds it",
			idleTimeout:          "24h",
			sessionExpiresAt:     timePtr(time.Now().Add(2 * time.Hour)),
			expectIdleExpiresAt:  true,
			expectCappedToJWTExp: true,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := s.setupConfig(currentTest.idleTimeout)

			// Create session and get token
			token, sessionID := s.createSessionWithToken(testUserID, cfg, currentTest.sessionExpiresAt)

			// Get the session before the request to capture original LastUsed
			sessionBefore, err := s.Storage.GetSessionPersister().Get(sessionID)
			s.Require().NoError(err)
			originalLastUsed := sessionBefore.LastUsed

			e := NewPublicRouter(cfg, s.Storage, nil, nil)

			requestBody := dto.ValidateSessionRequest{
				SessionToken: token,
			}
			bodyJson, err := json.Marshal(requestBody)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/sessions/validate", bytes.NewReader(bodyJson))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Equal(http.StatusOK, rec.Code)

			var response dto.ValidateSessionResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			s.Require().NoError(err)

			s.True(response.IsValid)
			s.NotNil(response.Claims)

			if currentTest.expectIdleExpiresAt {
				s.NotNil(response.IdleExpiresAt, "idle_expires_at should be set")

				if currentTest.expectCappedToJWTExp {
					// Should be capped to JWT expiration
					s.Equal(response.Claims.Expiration.Unix(), response.IdleExpiresAt.Unix(),
						"idle_expires_at should be capped to JWT expiration")
				} else {
					// Should be approximately now + idle timeout
					expectedTime := time.Now().Add(1 * time.Hour)
					s.InDelta(expectedTime.Unix(), response.IdleExpiresAt.Unix(), 5,
						"idle_expires_at should be now + idle timeout")
				}
			} else {
				s.Nil(response.IdleExpiresAt, "idle_expires_at should not be set when idle timeout is 0")
			}

			// Verify session LastUsed was updated
			sessionAfter, err := s.Storage.GetSessionPersister().Get(sessionID)
			s.Require().NoError(err)
			s.NotNil(sessionAfter)
			s.True(sessionAfter.LastUsed.After(originalLastUsed),
				"LastUsed should be updated to a newer timestamp (before: %v, after: %v)",
				originalLastUsed, sessionAfter.LastUsed)
		})
	}
}

func (s *sessionSuite) setupConfig(idleTimeout string) *config.Config {
	cfg := test.DefaultConfig
	cfg.Session.IdleTimeout = idleTimeout
	return &cfg
}

func (s *sessionSuite) createSessionWithCookie(userId uuid.UUID, cfg *config.Config, expiresAt *time.Time) (*http.Cookie, uuid.UUID) {
	manager := getDefaultSessionManager(s.Storage)

	userJWT := dto.UserJWT{
		UserID: userId.String(),
	}

	var token string
	var rawToken jwt.Token
	var err error

	if expiresAt != nil {
		token, rawToken, err = manager.GenerateJWT(userJWT, session.WithValue(jwt.ExpirationKey, expiresAt))
	} else {
		token, rawToken, err = manager.GenerateJWT(userJWT)
	}
	s.Require().NoError(err)

	sessionID, _ := rawToken.Get("session_id")
	sessionUUID := uuid.FromStringOrNil(sessionID.(string))

	session := models.Session{
		ID:        sessionUUID,
		UserID:    userId,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
		LastUsed:  time.Now().UTC(),
	}
	err = s.Storage.GetSessionPersister().Create(session)
	s.Require().NoError(err)

	cookie, err := manager.GenerateCookie(token)
	s.Require().NoError(err)

	return cookie, sessionUUID
}

func (s *sessionSuite) createSessionWithToken(userId uuid.UUID, cfg *config.Config, expiresAt *time.Time) (string, uuid.UUID) {
	manager := getDefaultSessionManager(s.Storage)

	userJWT := dto.UserJWT{
		UserID: userId.String(),
	}

	var token string
	var rawToken jwt.Token
	var err error

	if expiresAt != nil {
		token, rawToken, err = manager.GenerateJWT(userJWT, session.WithValue(jwt.ExpirationKey, expiresAt))
	} else {
		token, rawToken, err = manager.GenerateJWT(userJWT)
	}
	s.Require().NoError(err)

	sessionID, _ := rawToken.Get("session_id")
	sessionUUID := uuid.FromStringOrNil(sessionID.(string))

	session := models.Session{
		ID:        sessionUUID,
		UserID:    userId,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
		LastUsed:  time.Now().UTC(),
	}
	err = s.Storage.GetSessionPersister().Create(session)
	s.Require().NoError(err)

	return token, sessionUUID
}

func timePtr(t time.Time) *time.Time {
	return &t
}
