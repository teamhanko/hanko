package persistence

import (
	"embed"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/v2/config"
)

//go:embed migrations/*
var migrations embed.FS

// Persister is the persistence interface connecting to the database and capable of doing migrations
type persister struct {
	DB *pop.Connection
}

type Persister interface {
	GetAuditLogPersister() AuditLogPersister
	GetAuditLogPersisterWithConnection(tx *pop.Connection) AuditLogPersister
	GetConnection() *pop.Connection
	GetFlowPersister() FlowPersister
	GetFlowPersisterWithConnection(tx *pop.Connection) FlowPersister
	GetEmailPersister() EmailPersister
	GetEmailPersisterWithConnection(tx *pop.Connection) EmailPersister
	GetIdentityPersister() IdentityPersister
	GetIdentityPersisterWithConnection(tx *pop.Connection) IdentityPersister
	GetJwkPersister() JwkPersister
	GetJwkPersisterWithConnection(tx *pop.Connection) JwkPersister
	GetPasscodePersister() PasscodePersister
	GetPasscodePersisterWithConnection(tx *pop.Connection) PasscodePersister
	GetPasswordCredentialPersister() PasswordCredentialPersister
	GetPasswordCredentialPersisterWithConnection(tx *pop.Connection) PasswordCredentialPersister
	GetPrimaryEmailPersister() PrimaryEmailPersister
	GetPrimaryEmailPersisterWithConnection(tx *pop.Connection) PrimaryEmailPersister
	GetSamlCertificatePersister() SamlCertificatePersister
	GetSamlCertificatePersisterWithConnection(tx *pop.Connection) SamlCertificatePersister
	GetSamlStatePersister() SamlStatePersister
	GetSamlStatePersisterWithConnection(tx *pop.Connection) SamlStatePersister
	GetSamlIdentityPersister() SamlIdentityPersister
	GetSamlIdentityPersisterWithConnection(tx *pop.Connection) SamlIdentityPersister
	GetSamlIDPInitiatedRequestPersister() SamlIDPInitiatedRequestPersister
	GetSamlIDPInitiatedRequestPersisterWithConnection(tx *pop.Connection) SamlIDPInitiatedRequestPersister
	GetTokenPersister() TokenPersister
	GetTokenPersisterWithConnection(tx *pop.Connection) TokenPersister
	GetUserPersister() UserPersister
	GetUserPersisterWithConnection(tx *pop.Connection) UserPersister
	GetUserMetadataPersister() UserMetadataPersister
	GetUserMetadataPersisterWithConnection(tx *pop.Connection) UserMetadataPersister
	GetWebauthnCredentialPersister() WebauthnCredentialPersister
	GetWebauthnCredentialPersisterWithConnection(tx *pop.Connection) WebauthnCredentialPersister
	GetWebauthnSessionDataPersister() WebauthnSessionDataPersister
	GetWebauthnSessionDataPersisterWithConnection(tx *pop.Connection) WebauthnSessionDataPersister
	GetWebhookPersister(tx *pop.Connection) WebhookPersister
	GetTrustedDevicePersister() TrustedDevicePersister
	GetTrustedDevicePersisterWithConnection(tx *pop.Connection) TrustedDevicePersister
	GetUsernamePersister() UsernamePersister
	GetUsernamePersisterWithConnection(tx *pop.Connection) UsernamePersister
	GetSessionPersister() SessionPersister
	GetSessionPersisterWithConnection(tx *pop.Connection) SessionPersister
	GetOTPSecretPersister() OTPSecretPersister
	GetOTPSecretPersisterWithConnection(tx *pop.Connection) OTPSecretPersister
	GetWebauthnCredentialUserHandlePersister() WebauthnCredentialUserHandlePersister
	GetWebauthnCredentialUserHandlePersisterWithConnection(tx *pop.Connection) WebauthnCredentialUserHandlePersister
	GetTenantPersister() TenantPersister
	GetTenantPersisterWithConnection(tx *pop.Connection) TenantPersister
	Transaction(func(tx *pop.Connection) error) error
}

type Cleanup[T any] interface {
	FindExpired(cutoffTime time.Time, page, perPage int) ([]T, error)
	Delete(item T) error
}

type Migrator interface {
	MigrateUp() error
	MigrateDown(int) error
}

type Storage interface {
	Migrator
	Persister
}

// New return a new Persister Object with given configuration
func New(config config.Database) (Storage, error) {
	connectionDetails := &pop.ConnectionDetails{
		Pool:            5,
		IdlePool:        0,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 1 * time.Hour,
	}
	if len(config.Url) > 0 {
		connectionDetails.URL = config.Url
	} else {
		connectionDetails.Dialect = config.Dialect
		connectionDetails.Database = config.Database
		connectionDetails.Host = config.Host
		connectionDetails.Port = config.Port
		connectionDetails.User = config.User
		connectionDetails.Password = config.Password
	}

	DB, err := pop.NewConnection(connectionDetails)

	if err != nil {
		return nil, err
	}

	if err := DB.Open(); err != nil {
		return nil, err
	}

	return &persister{
		DB: DB,
	}, nil
}

// MigrateUp applies all pending up migrations to the Database
func (p *persister) MigrateUp() error {
	migrationBox, err := pop.NewMigrationBox(migrations, p.DB)
	if err != nil {
		return err
	}
	err = migrationBox.Up()
	if err != nil {
		return err
	}
	return nil
}

// MigrateDown migrates the Database down by the given number of steps
func (p *persister) MigrateDown(steps int) error {
	migrationBox, err := pop.NewMigrationBox(migrations, p.DB)
	if err != nil {
		return err
	}
	err = migrationBox.Down(steps)
	if err != nil {
		return err
	}
	return nil
}

func (p *persister) GetConnection() *pop.Connection {
	return p.DB
}

func (p *persister) GetFlowPersister() FlowPersister {
	return NewFlowPersister(p.DB)
}

func (p *persister) GetFlowPersisterWithConnection(tx *pop.Connection) FlowPersister {
	return NewFlowPersister(tx)
}

func (p *persister) GetIdentityPersister() IdentityPersister {
	return NewIdentityPersister(p.DB)
}

func (p *persister) GetIdentityPersisterWithConnection(tx *pop.Connection) IdentityPersister {
	return NewIdentityPersister(tx)
}

func (p *persister) GetUserPersister() UserPersister {
	return NewUserPersister(p.DB)
}

func (p *persister) GetUserPersisterWithConnection(tx *pop.Connection) UserPersister {
	return NewUserPersister(tx)
}

func (p *persister) GetPasscodePersister() PasscodePersister {
	return NewPasscodePersister(p.DB)
}

func (p *persister) GetPasscodePersisterWithConnection(tx *pop.Connection) PasscodePersister {
	return NewPasscodePersister(tx)
}

func (p *persister) GetPasswordCredentialPersister() PasswordCredentialPersister {
	return NewPasswordCredentialPersister(p.DB)
}

func (p *persister) GetPasswordCredentialPersisterWithConnection(tx *pop.Connection) PasswordCredentialPersister {
	return NewPasswordCredentialPersister(tx)
}

func (p *persister) GetTrustedDevicePersister() TrustedDevicePersister {
	return NewTrustedDevicePersister(p.DB)
}

func (p *persister) GetTrustedDevicePersisterWithConnection(tx *pop.Connection) TrustedDevicePersister {
	return NewTrustedDevicePersister(tx)
}

func (p *persister) GetUsernamePersister() UsernamePersister {
	return NewUsernamePersister(p.DB)
}

func (p *persister) GetUsernamePersisterWithConnection(tx *pop.Connection) UsernamePersister {
	return NewUsernamePersister(tx)
}

func (p *persister) GetOTPSecretPersister() OTPSecretPersister {
	return NewOTPSecretPersister(p.DB)
}

func (p *persister) GetOTPSecretPersisterWithConnection(tx *pop.Connection) OTPSecretPersister {
	return NewOTPSecretPersister(tx)
}

func (p *persister) GetWebauthnCredentialPersister() WebauthnCredentialPersister {
	return NewWebauthnCredentialPersister(p.DB)
}

func (p *persister) GetWebauthnCredentialPersisterWithConnection(tx *pop.Connection) WebauthnCredentialPersister {
	return NewWebauthnCredentialPersister(tx)
}

func (p *persister) GetWebauthnSessionDataPersister() WebauthnSessionDataPersister {
	return NewWebauthnSessionDataPersister(p.DB)
}

func (p *persister) GetWebauthnSessionDataPersisterWithConnection(tx *pop.Connection) WebauthnSessionDataPersister {
	return NewWebauthnSessionDataPersister(tx)
}

func (p *persister) GetJwkPersister() JwkPersister {
	return NewJwkPersister(p.DB)
}

func (p *persister) GetJwkPersisterWithConnection(tx *pop.Connection) JwkPersister {
	return NewJwkPersister(tx)
}

func (p *persister) GetAuditLogPersister() AuditLogPersister {
	return NewAuditLogPersister(p.DB)
}

func (p *persister) GetAuditLogPersisterWithConnection(tx *pop.Connection) AuditLogPersister {
	return NewAuditLogPersister(tx)
}

func (p *persister) GetEmailPersister() EmailPersister {
	return NewEmailPersister(p.DB)
}

func (p *persister) GetEmailPersisterWithConnection(tx *pop.Connection) EmailPersister {
	return NewEmailPersister(tx)
}

func (p *persister) GetPrimaryEmailPersister() PrimaryEmailPersister {
	return NewPrimaryEmailPersister(p.DB)
}

func (p *persister) GetPrimaryEmailPersisterWithConnection(tx *pop.Connection) PrimaryEmailPersister {
	return NewPrimaryEmailPersister(tx)
}

func (p *persister) Transaction(fn func(tx *pop.Connection) error) error {
	return p.DB.Transaction(fn)
}

func (p *persister) GetTokenPersister() TokenPersister {
	return NewTokenPersister(p.DB)
}

func (p *persister) GetTokenPersisterWithConnection(tx *pop.Connection) TokenPersister {
	return NewTokenPersister(tx)
}

func (p *persister) GetSamlStatePersister() SamlStatePersister {
	return NewSamlStatePersister(p.DB)
}

func (p *persister) GetSamlStatePersisterWithConnection(tx *pop.Connection) SamlStatePersister {
	return NewSamlStatePersister(tx)
}

func (p *persister) GetSamlCertificatePersister() SamlCertificatePersister {
	return NewSamlCertificatePersister(p.DB)
}

func (p *persister) GetSamlCertificatePersisterWithConnection(tx *pop.Connection) SamlCertificatePersister {
	return NewSamlCertificatePersister(tx)
}

func (p *persister) GetSamlIdentityPersister() SamlIdentityPersister {
	return NewSamlIdentityPersister(p.DB)
}

func (p *persister) GetSamlIdentityPersisterWithConnection(tx *pop.Connection) SamlIdentityPersister {
	return NewSamlIdentityPersister(tx)
}

func (p *persister) GetSamlIDPInitiatedRequestPersister() SamlIDPInitiatedRequestPersister {
	return NewSamlIDPInitiatedRequestPersister(p.DB)
}

func (p *persister) GetSamlIDPInitiatedRequestPersisterWithConnection(tx *pop.Connection) SamlIDPInitiatedRequestPersister {
	return NewSamlIDPInitiatedRequestPersister(tx)
}

func (p *persister) GetWebhookPersister(tx *pop.Connection) WebhookPersister {
	if tx != nil {
		return NewWebhookPersister(tx)
	}

	return NewWebhookPersister(p.DB)
}

func (p *persister) GetSessionPersister() SessionPersister {
	return NewSessionPersister(p.DB)
}

func (p *persister) GetSessionPersisterWithConnection(tx *pop.Connection) SessionPersister {
	return NewSessionPersister(tx)
}

func (p *persister) GetWebauthnCredentialUserHandlePersister() WebauthnCredentialUserHandlePersister {
	return NewWebauthnCredentialUserHandlePersister(p.DB)
}

func (p *persister) GetWebauthnCredentialUserHandlePersisterWithConnection(tx *pop.Connection) WebauthnCredentialUserHandlePersister {
	return NewWebauthnCredentialUserHandlePersister(tx)
}

func (p *persister) GetUserMetadataPersister() UserMetadataPersister {
	return NewUserMetadataPersister(p.DB)
}

func (p *persister) GetUserMetadataPersisterWithConnection(tx *pop.Connection) UserMetadataPersister {
	return NewUserMetadataPersister(tx)
}

func (p *persister) GetTenantPersister() TenantPersister {
	return NewTenantPersister(p.DB)
}

func (p *persister) GetTenantPersisterWithConnection(tx *pop.Connection) TenantPersister {
	return NewTenantPersister(tx)
}
