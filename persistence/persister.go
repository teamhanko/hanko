package persistence

import (
	"embed"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/config"
)

//go:embed migrations/*
var migrations embed.FS

// Persister is the persistence interface connecting to the database and capable of doing migrations
type Persister struct {
	DB                  *pop.Connection
	User                *UserPersister
	Passcode            *PasscodePersister
	WebAuthnCredential  *WebauthnCredentialPersister
	WebAuthnSessionData *WebauthnSessionDataPersister
	Jwk                 *JwkPersister
}

//New return a new Persister Object with given configuration
func New(config config.Database) (*Persister, error) {
	DB, err := pop.NewConnection(&pop.ConnectionDetails{
		Dialect:  config.Dialect,
		Database: config.Database,
		Host:     config.Host,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		Pool:     5,
		IdlePool: 0,
	})

	if err != nil {
		return nil, err
	}

	if err := DB.Open(); err != nil {
		return nil, err
	}

	return &Persister{
		DB:                  DB,
		User:                NewUserPersister(DB),
		Passcode:            NewPasscodePersister(DB),
		WebAuthnCredential:  NewWebauthnCredentialPersister(DB),
		WebAuthnSessionData: NewWebauthnSessionDataPersister(DB),
		Jwk:                 NewJwkPersister(DB),
	}, nil
}

// MigrateUp applies all pending up migrations to the Database
func (p *Persister) MigrateUp() error {
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
func (p *Persister) MigrateDown(steps int) error {
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
