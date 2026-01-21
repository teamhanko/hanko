package session

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v2/dto"
)

type Manager interface {
	GenerateJWT(user dto.UserJWT, opts ...JWTOptions) (string, jwt.Token, error)
	Verify(string) (jwt.Token, error)
	GenerateCookie(token string) (*http.Cookie, error)
	DeleteCookie() (*http.Cookie, error)
}

// Manager is used to create and verify session JWTs
type manager struct {
	jwtGenerator  jwk.Generator
	sessionLength time.Duration
	cookieConfig  cookieConfig
	issuer        string
	audience      []string
	jwtTemplate   *config.JWTTemplate
}

type cookieConfig struct {
	Name     string
	Domain   string
	HttpOnly bool
	SameSite http.SameSite
	Secure   bool
}

const (
	GeneratorCreateFailure = "failed to create session generator: %w"
)

// NewManager returns a new Manager which will be used to create and verify sessions JWTs
func NewManager(jwtGenerator jwk.Generator, config config.Config) (Manager, error) {
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

	return &manager{
		jwtGenerator:  jwtGenerator,
		sessionLength: duration,
		issuer:        config.Session.Issuer,
		cookieConfig: cookieConfig{
			Name:     config.Session.Cookie.GetName(),
			Domain:   config.Session.Cookie.Domain,
			HttpOnly: config.Session.Cookie.HttpOnly,
			SameSite: sameSite,
			Secure:   config.Session.Cookie.Secure,
		},
		audience:    audience,
		jwtTemplate: config.Session.JWTTemplate,
	}, nil
}

// GenerateJWT creates a new session JWT for the given user
func (m *manager) GenerateJWT(user dto.UserJWT, opts ...JWTOptions) (string, jwt.Token, error) {
	token := jwt.New()

	// Process the claim template if found
	if m.jwtTemplate != nil {
		if err := ProcessJWTTemplate(token, m.jwtTemplate.Claims, user); err != nil {
			return "", nil, err
		}
	}

	issuedAt := time.Now()
	expiration := issuedAt.Add(m.sessionLength)

	_ = token.Set(jwt.SubjectKey, user.UserID)
	_ = token.Set(jwt.IssuedAtKey, issuedAt)
	_ = token.Set(jwt.ExpirationKey, expiration)
	_ = token.Set(jwt.AudienceKey, m.audience)

	sessionID, err := uuid.NewV4()
	if err != nil {
		return "", nil, err
	}
	_ = token.Set("session_id", sessionID.String())

	if user.Email != nil {
		_ = token.Set("email", user.Email)
	}

	if user.Username != "" {
		_ = token.Set("username", user.Username)
	}

	if user.TenantID != nil {
		_ = token.Set("tenant_id", *user.TenantID)
	}

	for _, opt := range opts {
		opt(token)
	}

	if m.issuer != "" {
		_ = token.Set(jwt.IssuerKey, m.issuer)
	}

	signed, err := m.jwtGenerator.Sign(token)
	if err != nil {
		return "", nil, err
	}

	return string(signed), token, nil
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

// DeleteCookie returns a cookie that will expire the cookie on the frontend
func (m *manager) DeleteCookie() (*http.Cookie, error) {
	return &http.Cookie{
		Name:     m.cookieConfig.Name,
		Value:    "",
		Domain:   m.cookieConfig.Domain,
		Path:     "/",
		Secure:   m.cookieConfig.Secure,
		HttpOnly: m.cookieConfig.HttpOnly,
		SameSite: m.cookieConfig.SameSite,
		MaxAge:   -1,
	}, nil
}

type JWTOptions func(token jwt.Token)

func WithValue(key string, value interface{}) JWTOptions {
	return func(jwt jwt.Token) {
		_ = jwt.Set(key, value)
	}
}
