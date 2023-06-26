package oidc

import (
	"github.com/zitadel/oidc/v2/pkg/oidc"
	"github.com/zitadel/oidc/v2/pkg/op"
	"time"
)

var (
	// we use the default login UI and pass the (auth request) id
	defaultLoginURL = func(id string) string {
		return "/login/username?authRequestID=" + id
	}

	// clients to be used by the storage interface
	clients = map[string]*Client{}
)

// Client represents the storage model of an OAuth/OIDC client
type Client struct {
	id                             string
	secret                         string
	redirectURIs                   []string
	applicationType                op.ApplicationType
	authMethod                     oidc.AuthMethod
	loginURL                       func(string) string
	responseTypes                  []oidc.ResponseType
	grantTypes                     []oidc.GrantType
	accessTokenType                op.AccessTokenType
	devMode                        bool
	idTokenUserinfoClaimsAssertion bool
	clockSkew                      time.Duration
	postLogoutRedirectURIGlobs     []string
	redirectURIGlobs               []string

	keys map[string]interface{}
}

// GetID must return the client_id
func (c *Client) GetID() string {
	return c.id
}

// RedirectURIs must return the registered redirect_uris for Code and Implicit Flow
func (c *Client) RedirectURIs() []string {
	return c.redirectURIs
}

// PostLogoutRedirectURIs must return the registered post_logout_redirect_uris for sign-outs
func (c *Client) PostLogoutRedirectURIs() []string {
	return []string{}
}

// ApplicationType must return the type of the client (app, native, user agent)
func (c *Client) ApplicationType() op.ApplicationType {
	return c.applicationType
}

// AuthMethod must return the authentication method (client_secret_basic, client_secret_post, none, private_key_jwt)
func (c *Client) AuthMethod() oidc.AuthMethod {
	return c.authMethod
}

// ResponseTypes must return all allowed response types (code, id_token token, id_token)
// these must match with the allowed grant types
func (c *Client) ResponseTypes() []oidc.ResponseType {
	return c.responseTypes
}

// GrantTypes must return all allowed grant types (authorization_code, refresh_token, urn:ietf:params:oauth:grant-type:jwt-bearer)
func (c *Client) GrantTypes() []oidc.GrantType {
	return c.grantTypes
}

// LoginURL will be called to redirect the user (agent) to the login UI
// you could implement some logic here to redirect the users to different login UIs depending on the client
func (c *Client) LoginURL(id string) string {
	return c.loginURL(id)
}

// AccessTokenType must return the type of access token the client uses (Bearer (opaque) or JWT)
func (c *Client) AccessTokenType() op.AccessTokenType {
	return c.accessTokenType
}

// IDTokenLifetime must return the lifetime of the client's id_tokens
func (c *Client) IDTokenLifetime() time.Duration {
	return 1 * time.Hour
}

// DevMode enables the use of non-compliant configs such as redirect_uris (e.g. http schema for user agent client)
func (c *Client) DevMode() bool {
	return c.devMode
}

// RestrictAdditionalIdTokenScopes allows specifying which custom scopes shall be asserted into the id_token
func (c *Client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

// RestrictAdditionalAccessTokenScopes allows specifying which custom scopes shall be asserted into the JWT access_token
func (c *Client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

// IsScopeAllowed enables Client specific custom scopes validation
// in this example we allow the CustomScope for all clients
func (c *Client) IsScopeAllowed(scope string) bool {
	return false
}

// IDTokenUserinfoClaimsAssertion allows specifying if claims of scope profile, email, phone and address are asserted into the id_token
// even if an access token if issued which violates the OIDC Core spec
// (5.4. Requesting Claims using Scope Values: https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims)
// some clients though require that e.g. email is always in the id_token when requested even if an access_token is issued
func (c *Client) IDTokenUserinfoClaimsAssertion() bool {
	return c.idTokenUserinfoClaimsAssertion
}

// ClockSkew enables clients to instruct the OP to apply a clock skew on the various times and expirations
// (subtract from issued_at, add to expiration, ...)
func (c *Client) ClockSkew() time.Duration {
	return c.clockSkew
}

func (c *Client) GetKey(keyID string) (interface{}, bool) {
	k, ok := c.keys[keyID]
	return k, ok
}

// RegisterClients enables you to register clients for the example implementation
// there are some clients (web and native) to try out different cases
// add more if necessary
//
// RegisterClients should be called before the Storage is used so that there are
// no race conditions.
func RegisterClients(registerClients ...*Client) {
	for _, client := range registerClients {
		clients[client.id] = client
	}
}

// NativeClient will create a client of type native, which will always use PKCE and allow the use of refresh tokens
// user-defined redirectURIs may include:
// - http://localhost without port specification (e.g. http://localhost/auth/callback)
// - custom protocol (e.g. custom://auth/callback)
// (the examples will be used as default, if none is provided)
func NativeClient(id string, redirectURIs ...string) *Client {
	if len(redirectURIs) == 0 {
		redirectURIs = []string{
			"http://localhost/auth/callback",
			"custom://auth/callback",
		}
	}

	return &Client{
		id:                             id,
		secret:                         "", // no secret needed (due to PKCE)
		redirectURIs:                   redirectURIs,
		applicationType:                op.ApplicationTypeNative,
		authMethod:                     oidc.AuthMethodNone,
		loginURL:                       defaultLoginURL,
		responseTypes:                  []oidc.ResponseType{oidc.ResponseTypeCode},
		grantTypes:                     []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken},
		accessTokenType:                op.AccessTokenTypeBearer,
		devMode:                        false,
		idTokenUserinfoClaimsAssertion: false,
		clockSkew:                      0,
	}
}

// WebClient will create a client of type web, which will always use Basic Auth and allow the use of refresh tokens
// user-defined redirectURIs may include:
// - http://localhost with port specification (e.g. http://localhost:9999/auth/callback)
// (the example will be used as default, if none is provided)
func WebClient(id, secret string, redirectURIs ...string) *Client {
	if len(redirectURIs) == 0 {
		redirectURIs = []string{
			"http://localhost:9999/auth/callback",
		}
	}

	return &Client{
		id:                             id,
		secret:                         secret,
		redirectURIs:                   redirectURIs,
		applicationType:                op.ApplicationTypeWeb,
		authMethod:                     oidc.AuthMethodBasic,
		loginURL:                       defaultLoginURL,
		responseTypes:                  []oidc.ResponseType{oidc.ResponseTypeCode},
		grantTypes:                     []oidc.GrantType{oidc.GrantTypeCode, oidc.GrantTypeRefreshToken},
		accessTokenType:                op.AccessTokenTypeBearer,
		devMode:                        false,
		idTokenUserinfoClaimsAssertion: false,
		clockSkew:                      0,
	}
}

type hasRedirectGlobs struct {
	*Client
}

// RedirectURIGlobs provide wildcarding for additional valid redirects
func (c hasRedirectGlobs) RedirectURIGlobs() []string {
	return c.redirectURIGlobs
}

// PostLogoutRedirectURIGlobs provide extra wildcarding for additional valid redirects
func (c hasRedirectGlobs) PostLogoutRedirectURIGlobs() []string {
	return c.postLogoutRedirectURIGlobs
}

// RedirectGlobsClient wraps the client in a op.HasRedirectGlobs
// only if DevMode is enabled.
func RedirectGlobsClient(client *Client) op.Client {
	if client.devMode {
		return hasRedirectGlobs{client}
	}
	return client
}
