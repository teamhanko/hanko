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
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/session"
	"github.com/teamhanko/hanko/backend/v3/test"
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
	testTenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

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

			err = local_db.SyncSecretKeys(cfg, s.Storage)
			s.Require().NoError(err)

			// Create sess with cookie
			cookie, _ := s.createSessionWithCookie(testUserID, testTenantID, cfg, currentTest.sessionExpiresAt)

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
	testTenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

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

			err = local_db.SyncSecretKeys(cfg, s.Storage)
			s.Require().NoError(err)

			// Create session and get token
			token, sessionID := s.createSessionWithToken(testUserID, testTenantID, cfg, currentTest.sessionExpiresAt)

			// Get the session before the request to capture original LastUsed
			sessionBefore, err := s.Storage.GetSessionPersister().Get(sessionID, testTenantID)
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
			sessionAfter, err := s.Storage.GetSessionPersister().Get(sessionID, testTenantID)
			s.Require().NoError(err)
			s.NotNil(sessionAfter)
			s.True(sessionAfter.LastUsed.After(originalLastUsed),
				"LastUsed should be updated to a newer timestamp (before: %v, after: %v)",
				originalLastUsed, sessionAfter.LastUsed)
		})
	}
}

func (s *sessionSuite) TestSessionHandler_ValidateSession_Organizations() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	testTenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name              string
		userID            string
		expectedOrgID     string
		expectedOrgName   string
		expectedRoleSlugs []string
	}{
		{
			name:              "member with a role binding",
			userID:            "ec4ef049-5b88-4321-a173-21b0eff06a04",
			expectedOrgID:     "bebebebe-0000-0000-0000-000000000001",
			expectedOrgName:   "Session Test Org",
			expectedRoleSlugs: []string{"admin"},
		},
		{
			name:   "user with no organization memberships",
			userID: "38bf5a00-d7ea-40a5-a5de-48722c148925",
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			cfg := s.setupConfig("0s")

			err = local_db.SyncSecretKeys(cfg, s.Storage)
			s.Require().NoError(err)

			cookie, _ := s.createSessionWithCookie(uuid.FromStringOrNil(currentTest.userID), testTenantID, cfg, nil)

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

			if currentTest.expectedOrgID == "" {
				s.Empty(response.Organizations)
				return
			}

			s.Require().Len(response.Organizations, 1)
			s.Equal(uuid.FromStringOrNil(currentTest.expectedOrgID), response.Organizations[0].ID)
			s.Equal(currentTest.expectedOrgName, response.Organizations[0].Name)
			s.Equal(currentTest.expectedRoleSlugs, response.Organizations[0].Roles)
		})
	}
}

func (s *sessionSuite) TestSessionHandler_ValidateSession_Organizations_MultiTenantRouting() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	testTenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	cfg := test.DefaultConfig
	cfg.MultiTenancy.Enabled = true
	err = cfg.PostProcess()
	s.Require().NoError(err)
	cfg.Session.IdleTimeout = "0s"

	err = generateSigningKeyForTenant(s.Storage, testTenantID)
	s.Require().NoError(err)

	cookie, _ := s.createSessionWithCookie(uuid.FromStringOrNil("ec4ef049-5b88-4321-a173-21b0eff06a04"), testTenantID, &cfg, nil)

	e := NewPublicRouter(&cfg, s.Storage, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/00000000-0000-0000-0000-000000000001/sessions/validate", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var response dto.ValidateSessionResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	s.Require().NoError(err)

	s.True(response.IsValid)
	s.Require().Len(response.Organizations, 1)
	s.Equal(uuid.FromStringOrNil("bebebebe-0000-0000-0000-000000000001"), response.Organizations[0].ID)
	s.Equal("Session Test Org", response.Organizations[0].Name)
	s.Equal([]string{"admin"}, response.Organizations[0].Roles)
}

func (s *sessionSuite) TestSessionHandler_ValidateSessionFromBody_Organizations() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	testUserID := uuid.FromStringOrNil("ec4ef049-5b88-4321-a173-21b0eff06a04")
	testTenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	cfg := s.setupConfig("0s")

	err = local_db.SyncSecretKeys(cfg, s.Storage)
	s.Require().NoError(err)

	token, _ := s.createSessionWithToken(testUserID, testTenantID, cfg, nil)

	e := NewPublicRouter(cfg, s.Storage, nil, nil)

	requestBody := dto.ValidateSessionRequest{SessionToken: token}
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
	s.Require().Len(response.Organizations, 1)
	s.Equal(uuid.FromStringOrNil("bebebebe-0000-0000-0000-000000000001"), response.Organizations[0].ID)
	s.Equal("Session Test Org", response.Organizations[0].Name)
	s.Equal([]string{"admin"}, response.Organizations[0].Roles)
}

func (s *sessionSuite) TestSessionHandler_ValidateSession_Organizations_RevokedMembershipTakesEffectImmediately() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/sessions")
	s.Require().NoError(err)

	testUserID := uuid.FromStringOrNil("ec4ef049-5b88-4321-a173-21b0eff06a04")
	testTenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	cfg := s.setupConfig("0s")
	err = local_db.SyncSecretKeys(cfg, s.Storage)
	s.Require().NoError(err)

	cookie, _ := s.createSessionWithCookie(testUserID, testTenantID, cfg, nil)
	e := NewPublicRouter(cfg, s.Storage, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/sessions/validate", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var before dto.ValidateSessionResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &before))
	s.Require().Len(before.Organizations, 1)

	// Revoke the membership - without reissuing the session token.
	membership, err := s.Storage.GetOrganizationMembershipPersister().Get(
		testUserID, uuid.FromStringOrNil("bebebebe-0000-0000-0000-000000000001"), testTenantID)
	s.Require().NoError(err)
	s.Require().NotNil(membership)
	err = s.Storage.GetOrganizationMembershipPersister().Delete(*membership)
	s.Require().NoError(err)

	req2 := httptest.NewRequest(http.MethodGet, "/sessions/validate", nil)
	req2.AddCookie(cookie)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	var after dto.ValidateSessionResponse
	s.Require().NoError(json.Unmarshal(rec2.Body.Bytes(), &after))
	s.Empty(after.Organizations, "revoking a membership should take effect on the very next validation call, without reissuing the session token")
}

func (s *sessionSuite) setupConfig(idleTimeout string) *config.Config {
	cfg := test.DefaultConfig
	err := cfg.PostProcess()
	s.Require().NoError(err)

	cfg.Session.IdleTimeout = idleTimeout
	return &cfg
}

func (s *sessionSuite) createSessionWithCookie(userId uuid.UUID, tenantID uuid.UUID, cfg *config.Config, expiresAt *time.Time) (*http.Cookie, uuid.UUID) {
	manager := getDefaultSessionManager(s.Storage)

	userJWT := dto.UserJWT{
		UserID: userId.String(),
	}

	var token string
	var rawToken jwt.Token
	var err error

	if expiresAt != nil {
		token, rawToken, err = manager.GenerateJWT(userJWT, tenantID, session.WithValue(jwt.ExpirationKey, expiresAt))
	} else {
		token, rawToken, err = manager.GenerateJWT(userJWT, tenantID)
	}
	s.Require().NoError(err)

	sessionID, _ := rawToken.Get("session_id")
	sessionUUID := uuid.FromStringOrNil(sessionID.(string))

	session := models.Session{
		ID:        sessionUUID,
		TenantID:  tenantID,
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

func (s *sessionSuite) createSessionWithToken(userId uuid.UUID, tenantID uuid.UUID, cfg *config.Config, expiresAt *time.Time) (string, uuid.UUID) {
	manager := getDefaultSessionManager(s.Storage)

	userJWT := dto.UserJWT{
		UserID: userId.String(),
	}

	var token string
	var rawToken jwt.Token
	var err error

	if expiresAt != nil {
		token, rawToken, err = manager.GenerateJWT(userJWT, tenantID, session.WithValue(jwt.ExpirationKey, expiresAt))
	} else {
		token, rawToken, err = manager.GenerateJWT(userJWT, tenantID)
	}
	s.Require().NoError(err)

	sessionID, _ := rawToken.Get("session_id")
	sessionUUID := uuid.FromStringOrNil(sessionID.(string))

	session := models.Session{
		ID:        sessionUUID,
		TenantID:  tenantID,
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
