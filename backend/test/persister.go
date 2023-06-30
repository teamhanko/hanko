package test

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewPersister(user []models.User, passcodes []models.Passcode, jwks []models.Jwk, credentials []models.WebauthnCredential, sessionData []models.WebauthnSessionData, passwords []models.PasswordCredential, auditLogs []models.AuditLog, emails []models.Email, primaryEmails []models.PrimaryEmail, identities []models.Identity, tokens []models.Token, accessTokens []models.AccessToken, refreshTokens []models.RefreshToken, keys []models.Key, authRequests []models.AuthRequest, codes map[string]uuid.UUID) persistence.Persister {
	return &persister{
		userPersister:                NewUserPersister(user),
		passcodePersister:            NewPasscodePersister(passcodes),
		jwkPersister:                 NewJwkPersister(jwks),
		webauthnCredentialPersister:  NewWebauthnCredentialPersister(credentials),
		webauthnSessionDataPersister: NewWebauthnSessionDataPersister(sessionData),
		passwordCredentialPersister:  NewPasswordCredentialPersister(passwords),
		auditLogPersister:            NewAuditLogPersister(auditLogs),
		emailPersister:               NewEmailPersister(emails),
		primaryEmailPersister:        NewPrimaryEmailPersister(primaryEmails),
		identityPersister:            NewIdentityPersister(identities),
		tokenPersister:               NewTokenPersister(tokens),
		oidcAccessTokensPersister:    NewOidcAccessTokensPersister(accessTokens),
		oidcRefreshTokensPersister:   NewOidcRefreshTokensPersister(refreshTokens),
		oidcKeysPersister:            NewOidcKeysPersister(keys),
		oidcAuthRequestsPersister:    NewOidcAuthRequestsPersister(authRequests, codes),
	}
}

type persister struct {
	userPersister                persistence.UserPersister
	passcodePersister            persistence.PasscodePersister
	jwkPersister                 persistence.JwkPersister
	webauthnCredentialPersister  persistence.WebauthnCredentialPersister
	webauthnSessionDataPersister persistence.WebauthnSessionDataPersister
	passwordCredentialPersister  persistence.PasswordCredentialPersister
	auditLogPersister            persistence.AuditLogPersister
	emailPersister               persistence.EmailPersister
	primaryEmailPersister        persistence.PrimaryEmailPersister
	identityPersister            persistence.IdentityPersister
	tokenPersister               persistence.TokenPersister
	oidcAccessTokensPersister    persistence.OIDCAccessTokenPersister
	oidcRefreshTokensPersister   persistence.OIDCRefreshTokenPersister
	oidcKeysPersister            persistence.OIDCKeyPersister
	oidcAuthRequestsPersister    persistence.OIDCAuthRequestPersister
}

func (p *persister) GetPasswordCredentialPersister() persistence.PasswordCredentialPersister {
	return p.passwordCredentialPersister
}

func (p *persister) GetPasswordCredentialPersisterWithConnection(tx *pop.Connection) persistence.PasswordCredentialPersister {
	return p.passwordCredentialPersister
}

func (p *persister) GetConnection() *pop.Connection {
	return nil
}

func (p *persister) Transaction(fn func(tx *pop.Connection) error) error {
	return fn(nil)
}

func (p *persister) GetUserPersister() persistence.UserPersister {
	return p.userPersister
}

func (p *persister) GetUserPersisterWithConnection(tx *pop.Connection) persistence.UserPersister {
	return p.userPersister
}

func (p *persister) GetPasscodePersister() persistence.PasscodePersister {
	return p.passcodePersister
}

func (p *persister) GetPasscodePersisterWithConnection(tx *pop.Connection) persistence.PasscodePersister {
	return p.passcodePersister
}

func (p *persister) GetWebauthnCredentialPersister() persistence.WebauthnCredentialPersister {
	return p.webauthnCredentialPersister
}

func (p *persister) GetWebauthnCredentialPersisterWithConnection(tx *pop.Connection) persistence.WebauthnCredentialPersister {
	return p.webauthnCredentialPersister
}

func (p *persister) GetWebauthnSessionDataPersister() persistence.WebauthnSessionDataPersister {
	return p.webauthnSessionDataPersister
}

func (p *persister) GetWebauthnSessionDataPersisterWithConnection(tx *pop.Connection) persistence.WebauthnSessionDataPersister {
	return p.webauthnSessionDataPersister
}

func (p *persister) GetJwkPersister() persistence.JwkPersister {
	return p.jwkPersister
}

func (p *persister) GetJwkPersisterWithConnection(tx *pop.Connection) persistence.JwkPersister {
	return p.jwkPersister
}

func (p *persister) GetAuditLogPersister() persistence.AuditLogPersister {
	return p.auditLogPersister
}

func (p *persister) GetAuditLogPersisterWithConnection(tx *pop.Connection) persistence.AuditLogPersister {
	return p.auditLogPersister
}

func (p *persister) GetEmailPersister() persistence.EmailPersister {
	return p.emailPersister
}

func (p *persister) GetEmailPersisterWithConnection(tx *pop.Connection) persistence.EmailPersister {
	return p.emailPersister
}

func (p *persister) GetPrimaryEmailPersister() persistence.PrimaryEmailPersister {
	return p.primaryEmailPersister
}

func (p *persister) GetPrimaryEmailPersisterWithConnection(tx *pop.Connection) persistence.PrimaryEmailPersister {
	return p.primaryEmailPersister
}

func (p *persister) GetIdentityPersister() persistence.IdentityPersister {
	return p.identityPersister
}

func (p *persister) GetIdentityPersisterWithConnection(tx *pop.Connection) persistence.IdentityPersister {
	return p.identityPersister

}

func (p *persister) GetTokenPersister() persistence.TokenPersister {
	return p.tokenPersister
}

func (p *persister) GetTokenPersisterWithConnection(tx *pop.Connection) persistence.TokenPersister {
	return p.tokenPersister
}

func (p *persister) GetOIDCAccessTokenPersister() persistence.OIDCAccessTokenPersister {
	return p.oidcAccessTokensPersister
}

func (p *persister) GetOIDCAccessTokenPersisterWithConnection(tx *pop.Connection) persistence.OIDCAccessTokenPersister {
	return p.oidcAccessTokensPersister
}

func (p *persister) GetOIDCRefreshTokenPersister() persistence.OIDCRefreshTokenPersister {
	return p.oidcRefreshTokensPersister
}

func (p *persister) GetOIDCRefreshTokenPersisterWithConnection(tx *pop.Connection) persistence.OIDCRefreshTokenPersister {
	return p.oidcRefreshTokensPersister
}

func (p *persister) GetOIDCKeyPersister() persistence.OIDCKeyPersister {
	return p.oidcKeysPersister
}

func (p *persister) GetOIDCKeyPersisterWithConnection(tx *pop.Connection) persistence.OIDCKeyPersister {
	return p.oidcKeysPersister
}

func (p *persister) GetOIDCAuthRequestPersister() persistence.OIDCAuthRequestPersister {
	return p.oidcAuthRequestsPersister
}

func (p *persister) GetOIDCAuthRequestPersisterWithConnection(tx *pop.Connection) persistence.OIDCAuthRequestPersister {
	return p.oidcAuthRequestsPersister
}
