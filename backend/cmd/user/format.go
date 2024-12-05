package user

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/invopop/jsonschema"
	"time"
)

// ImportOrExportEmail The import/export format for a user's email
type ImportOrExportEmail struct {
	// Address Valid email address
	Address string `json:"address" yaml:"address" jsonschema:"format=email" validate:"email"`
	// IsPrimary indicates if this is the primary email of the users. In the Emails array there has to be exactly one primary email.
	IsPrimary bool `json:"is_primary" yaml:"is_primary"`
	// IsVerified indicates if the email address was previously verified.
	IsVerified bool `json:"is_verified" yaml:"is_verified"`
}

func (ImportOrExportEmail) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Title = "ImportEmail"
}

// Emails Array of email addresses
type Emails []ImportOrExportEmail

type ImportWebauthnCredential struct {
	// ID of the webauthn credential.
	ID string `json:"id" yaml:"id" validate:"required"`
	// Optional Name of the webauthn credential.
	Name *string `json:"name" yaml:"name" validate:"omitempty"`
	// The PublicKey of the credential.
	PublicKey string `json:"public_key" yaml:"public_key" validate:"required"`
	// The AttestationType the credential was created with.
	AttestationType string `json:"attestation_type" yaml:"attestation_type" validate:"required"`
	// Optional AAGUID of the authenticator on which the credential was created on.
	AAGUID uuid.UUID `json:"aaguid" yaml:"aaguid" validate:"omitempty,uuid4"`
	// Optional SignCount of the webauthn credential.
	SignCount int `json:"sign_count" yaml:"sign_count"`
	// LastUsedAt optional timestamp when the webauthn credential was last used.
	LastUsedAt *time.Time `json:"last_used_at" yaml:"last_used_at" validate:"omitempty"`
	// CreatedAt optional timestamp of the webauthn credentials' creation. Will be set to the import date if not provided.
	CreatedAt *time.Time `json:"created_at" yaml:"created_at" validate:"omitempty"`
	// UpdatedAt optional timestamp of the last update to the webauthn credential. Will be set to the import date if not provided.
	UpdatedAt *time.Time `json:"updated_at" yaml:"updated_at" validate:"omitempty"`
	// Optional list of supported Transports by the authenticator.
	Transports []string `json:"transports" yaml:"transports" validate:"omitempty,unique"`
	// BackupEligible flag indicates if the webauthn credential can be backed up (e.g. in Apple KeyChain, ...). If the information is not available set it to false.
	BackupEligible bool `json:"backup_eligible" yaml:"backup_eligible"`
	// BackupState flag indicates if the webauthn credential is backed up (e.g. in Apple KeyChain, ...). If the information is not available set it to false.
	BackupState bool `json:"backup_state" yaml:"backup_state"`
	// MFAOnly flag indicates if the webauthn credential can only be used in combination with another login factor (e.g. password, ...).
	MFAOnly bool `json:"mfa_only" yaml:"mfa_only"`
	// UserHandle optional user id which was used to create the credential with.
	// Populate only when user id was not an uuid v4 and the webauthn credential is not an MFAOnly credential.
	UserHandle *string `json:"user_handle" yaml:"user_handle" validate:"omitempty,excluded_if=MFAOnly true"`
}

type ImportWebauthnCredentials []ImportWebauthnCredential

type ImportPasswordCredential struct {
	// Password hash of the password in bcrypt format.
	Password string `json:"password" yaml:"password" validate:"required,startswith=$2a$"`
	// CreatedAt optional timestamp when the password was created. Will be set to the import date if not provided.
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"created_at" validate:"omitempty"`
	// UpdatedAt optional timestamp of the last update to the password. Will be set to the import date if not provided.
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"updated_at" validate:"omitempty"`
}

type ImportOTPSecret struct {
	// Secret of the TOTP credential. TOTP credential must be generated for a period of 30 seconds and SHA1 hash algorithm.
	Secret string `json:"secret" yaml:"secret" validate:"required"`
	// CreatedAt optional timestamp when the otp secret was created. Will be set to the import date if not provided.
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"created_at" validate:"omitempty"`
	// UpdatedAt optional timestamp of the last update to the otp secret. Will be set to the import date if not provided.
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"updated_at" validate:"omitempty"`
}

// ImportOrExportEntry represents a user to be imported/export to the Hanko database
type ImportOrExportEntry struct {
	// UserID optional uuid.v4. If not provided a new one will be generated for the user
	UserID string `json:"user_id,omitempty" yaml:"user_id" validate:"omitempty,uuid4"`
	// Emails optional list of emails
	Emails Emails `json:"emails" yaml:"emails" jsonschema:"type=array,minItems=1" validate:"required_if=Username 0,unique=Address,dive"`
	// Username optional username of the user
	Username *string `json:"username,omitempty" yaml:"username" validate:"required_if=Emails 0,omitempty,gte=1"`
	// WebauthnCredentials optional list of webauthn credentials of a user. Includes passkeys and MFA credentials.
	WebauthnCredentials ImportWebauthnCredentials `json:"webauthn_credentials,omitempty" yaml:"webauthn_credentials" validate:"omitempty,unique=ID,dive"`
	// Password optional password.
	Password *ImportPasswordCredential `json:"password" yaml:"password" validate:"omitempty"`
	// OTPSecret optional TOTP secret for MFA.
	OTPSecret *ImportOTPSecret `json:"otp_secret" yaml:"otp_secret" validate:"omitempty"`
	// CreatedAt optional timestamp of the users' creation. Will be set to the import date if not provided.
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"created_at" validate:"omitempty"`
	// UpdatedAt optional timestamp of the last update to the user. Will be set to the import date if not provided.
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"updated_at" validate:"omitempty"`
}

func (ImportOrExportEntry) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Title = "ImportEntry"
}

// ImportOrExportList a list of ImportEntries
type ImportOrExportList []ImportOrExportEntry

func (ImportOrExportList) JSONSchemaExtend(schema *jsonschema.Schema) {
	date := time.Date(2024, 8, 17, 12, 5, 15, 651387237, time.UTC)
	username := "example"
	schema.Examples = []any{
		[]ImportOrExportEntry{
			{
				UserID: "a9ae6bc8-d829-43de-b672-f50230833877",
				Emails: Emails{
					{"test@example.com", true, true},
					{"test+1@example.com", false, false},
				},
				CreatedAt: &date,
				UpdatedAt: &date,
			},
			{
				UserID: "2f0649cf-c71e-48a5-92c3-210addb80281",
				Emails: Emails{
					{"test2@example.com", true, true},
					{"test2+1@example.com", false, false},
				},
				CreatedAt: &date,
				UpdatedAt: &date,
			},
		},
		[]ImportOrExportEntry{
			{
				Username: &username,
				Password: &ImportPasswordCredential{
					Password:  "$2a$12$mFbud0mLsD/q.WG7/9pNQemlAHs3H4o8zAv44gsUF1v1awsdqTh7.",
					CreatedAt: &date,
					UpdatedAt: &date,
				},
			},
		},
	}
}

func (entry *ImportOrExportEntry) validate(v *validator.Validate) error {
	err := v.Struct(entry)
	if err != nil {
		return err
	}
	primaryEmailAddresses := 0
	for _, email := range entry.Emails {
		if email.IsPrimary {
			primaryEmailAddresses++
		}
	}

	if len(entry.Emails) > 0 && primaryEmailAddresses != 1 {
		return errors.New(fmt.Sprintf("Need exactly one primary email, got %v", primaryEmailAddresses))
	}

	return nil
}
