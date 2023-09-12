package session

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	hankoJwk "github.com/teamhanko/hanko/backend/crypto/jwk"
	hankoJwt "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"net/http"
	"time"
)

type Manager interface {
	GenerateJWT(uuid.UUID) (string, error)
	Verify(string) (jwt.Token, error)
	GenerateCookie(string) (*http.Cookie, error)
	GenerateCookieOrHeader(uuid.UUID, echo.Context) error
	ExchangeRefreshToken(string, echo.Context) error
	DeleteCookie(echo.Context) error
}

// Manager is used to create and verify session JWTs
type manager struct {
	jwtGenerator       hankoJwt.Generator
	sessionLength      time.Duration
	cookieConfig       cookieConfig
	enableHeader       bool
	enableRefreshToken bool
	refreshTokenPath   string
	issuer             string
	audience           []string
	persister          persistence.SessionPersister
}

type cookieConfig struct {
	Name     string
	Domain   string
	HttpOnly bool
	SameSite http.SameSite
	Secure   bool
}

// NewManager returns a new Manager which will be used to create and verify sessions JWTs
func NewManager(jwkManager hankoJwk.Manager, config config.Config, persister persistence.SessionPersister) (Manager, error) {
	signatureKey, err := jwkManager.GetSigningKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	verificationKeys, err := jwkManager.GetPublicKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	g, err := hankoJwt.NewGenerator(signatureKey, verificationKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}

	duration, _ := time.ParseDuration(config.Session.Lifespan) // error can be ignored, value is checked in config validation
	sameSite := http.SameSite(0)
	switch config.Session.Cookie.SameSite {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}
	var audience []string
	if config.Session.Audience != nil && len(config.Session.Audience) > 0 {
		audience = config.Session.Audience
	} else {
		audience = []string{config.Webauthn.RelyingParty.Id}
	}

	refreshTokenPath := "/session/exchange"
	if len(config.Server.Public.PathPrefix) > 0 {
		refreshTokenPath = config.Server.Public.PathPrefix + refreshTokenPath
	}

	return &manager{
		jwtGenerator:  g,
		sessionLength: duration,
		issuer:        config.Session.Issuer,
		cookieConfig: cookieConfig{
			Name:     config.Session.Cookie.GetName(),
			Domain:   config.Session.Cookie.Domain,
			HttpOnly: config.Session.Cookie.HttpOnly,
			SameSite: sameSite,
			Secure:   config.Session.Cookie.Secure,
		},
		enableHeader:       config.Session.EnableAuthTokenHeader,
		enableRefreshToken: config.Session.EnableRefreshToken,
		refreshTokenPath:   refreshTokenPath,
		audience:           audience,
		persister:          persister,
	}, nil
}

// GenerateJWT creates a new session JWT for the given user
func (m *manager) GenerateJWT(userId uuid.UUID) (string, error) {
	issuedAt := time.Now()
	expiration := issuedAt.Add(m.sessionLength)

	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, userId.String())
	_ = token.Set(jwt.IssuedAtKey, issuedAt)
	_ = token.Set(jwt.ExpirationKey, expiration)
	_ = token.Set(jwt.AudienceKey, m.audience)
	if m.issuer != "" {
		_ = token.Set(jwt.IssuerKey, m.issuer)
	}

	signed, err := m.jwtGenerator.Sign(token)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

// Verify verifies the given JWT and returns a parsed one if verification was successful
func (m *manager) Verify(token string) (jwt.Token, error) {
	parsedToken, err := m.jwtGenerator.Verify([]byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to verify session token: %w", err)
	}

	return parsedToken, nil
}

// GenerateCookie creates a new session cookie for the given user
func (m *manager) GenerateCookie(token string) (*http.Cookie, error) {
	return &http.Cookie{
		Name:     m.cookieConfig.Name,
		Value:    token,
		Domain:   m.cookieConfig.Domain,
		Path:     "/",
		Secure:   m.cookieConfig.Secure,
		HttpOnly: m.cookieConfig.HttpOnly,
		SameSite: m.cookieConfig.SameSite,
		MaxAge:   int(m.sessionLength.Seconds()),
	}, nil
}

// GenerateRefreshCookie creates a new refresh cookie for the given user
func (m *manager) GenerateRefreshCookie(sessionId string) (*http.Cookie, error) {
	return &http.Cookie{
		Name:     m.cookieConfig.Name + "-refresh",
		Value:    sessionId,
		Domain:   m.cookieConfig.Domain,
		Path:     m.refreshTokenPath,
		Secure:   m.cookieConfig.Secure,
		HttpOnly: m.cookieConfig.HttpOnly,
		SameSite: m.cookieConfig.SameSite,
	}, nil
}

// GenerateCookieOrHeader creates a new session cookie or applies the header for the given user
func (m *manager) GenerateCookieOrHeader(userId uuid.UUID, e echo.Context) error {
	token, err := m.GenerateJWT(userId)
	if err != nil {
		return err
	}

	if m.enableHeader {
		e.Response().Header().Set("X-Auth-Token", token)
	} else {
		cookie, _ := m.GenerateCookie(token)
		e.SetCookie(cookie)
	}

	e.Response().Header().Set("X-Session-Lifetime", fmt.Sprintf("%d", int(m.sessionLength.Seconds())))

	if !m.enableRefreshToken || m.persister == nil {
		return nil
	}

	session, err := models.NewSession(userId)
	if err != nil {
		return err
	}

	err = m.persister.Create(*session)
	if err != nil {
		return err
	}

	if m.enableHeader {
		e.Response().Header().Set("X-Refresh-Token", session.ID)
	} else {
		cookie, _ := m.GenerateRefreshCookie(session.ID)
		e.SetCookie(cookie)
	}

	return nil
}

// ExchangeRefreshToken refreshes the session cookie for the given user based on the given id of the refresh token
func (m *manager) ExchangeRefreshToken(id string, e echo.Context) error {
	sess, err := m.persister.Get(id)
	if err != nil {
		return err
	}

	if sess == nil {
		return errors.New("session not found")
	}

	err = m.persister.Delete(id)
	if err != nil {
		return err
	}

	return m.GenerateCookieOrHeader(sess.UserID, e)
}

// DeleteCookie returns a cookie that will expire the cookie on the frontend
func (m *manager) DeleteCookie(e echo.Context) error {
	if m.enableHeader {
		return nil
	}

	e.SetCookie(&http.Cookie{
		Name:     m.cookieConfig.Name,
		Value:    "",
		Domain:   m.cookieConfig.Domain,
		Path:     "/",
		Secure:   m.cookieConfig.Secure,
		HttpOnly: m.cookieConfig.HttpOnly,
		SameSite: m.cookieConfig.SameSite,
		MaxAge:   -1,
	})

	return nil
}
