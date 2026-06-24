package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestUserSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(userSuite))
}

type userSuite struct {
	test.Suite
}

func (s *userSuite) TestUserHandler_Me() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	err := s.LoadFixtures("../test/fixtures/user_with_webauthn_credential")
	s.Require().NoError(err)

	userId := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")
	tenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	cfg := test.DefaultConfig
	cfg.ThirdParty.Providers.Google = config.ThirdPartyProvider{
		Enabled:     true,
		ID:          "google",
		DisplayName: "Google",
	}
	err = cfg.PostProcess()
	s.Require().NoError(err)

	err = local_db.SyncSecretKeys(&cfg, s.Storage)
	s.Require().NoError(err)

	cfg.Passkey.Enabled = true
	cfg.MFA.Enabled = true
	cfg.MFA.TOTP.Enabled = true
	cfg.MFA.SecurityKeys.Enabled = true
	e := NewPublicRouter(&cfg, s.Storage, nil, nil)

	cookie, err := generateSessionCookie(s.Storage, userId, tenantID)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusOK, rec.Code) {
		response := dto.ProfileData{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		s.NoError(err)
		s.Equal(userId, response.UserID)
		s.Len(response.Emails, 1)
		s.Equal("john.doe@example.com", response.Emails[0].Address)
		s.True(response.Emails[0].IsVerified)
		s.Len(response.Passkeys, 1)
		s.Equal("P8fcQ6U8zxJRzhI0yuUCOxcA_UyAs0jbauO5ektj4SM", response.Passkeys[0].ID)
		s.Len(response.SecurityKeys, 1)
		s.Equal("security-key-cred-id", response.SecurityKeys[0].ID)
		s.False(response.MFAConfig.AuthAppSetUp)
		s.True(response.MFAConfig.TOTPEnabled)
		s.True(response.MFAConfig.SecurityKeysEnabled)
		s.NotNil(response.Username)
		s.Equal("johndoe", response.Username.Username)
		s.Equal("John Doe", response.Name)
		s.Equal("John", response.GivenName)
		s.Equal("Doe", response.FamilyName)
		s.Equal("https://example.com/john.jpg", response.Picture)
		s.NotNil(response.Metadata)
		s.Contains(string(response.Metadata.Public), "tester")
		s.Contains(string(response.Metadata.Unsafe), "debug")
		s.NotContains(string(rec.Body.Bytes()), "private_metadata")
		s.NotContains(string(rec.Body.Bytes()), "quota")
		s.Len(response.Identities, 1)
		s.Equal("Google", response.Identities[0].Provider)
	}
}

func (s *userSuite) TestUserHandler_Logout() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	err := s.LoadFixtures("../test/fixtures/user")
	s.Require().NoError(err)

	userId := uuid.FromStringOrNil("b5dd5267-b462-48be-b70d-bcd6f1bbe7a5")
	tenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	cfg := test.DefaultConfig
	err = cfg.PostProcess()
	s.Require().NoError(err)

	err = local_db.SyncSecretKeys(&cfg, s.Storage)
	s.Require().NoError(err)

	e := NewPublicRouter(&cfg, s.Storage, nil, nil)

	cookie, err := generateSessionCookie(s.Storage, userId, tenantID)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if s.Equal(http.StatusNoContent, rec.Code) {
		cookie := rec.Header().Get("Set-Cookie")
		s.NotEmpty(cookie)

		split := strings.Split(cookie, ";")
		s.Equal("Max-Age=0", strings.TrimSpace(split[2]))
	}
}
