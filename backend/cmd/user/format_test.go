package user

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const validUUID = "62418053-a2cd-47a8-9b61-4426380d263a"
const invalidUUID = "notvalid"

func TestImportEntry_validate(t *testing.T) {
	v := validator.New()
	validUsername := "example"
	emptyUsername := ""
	type fields struct {
		UserID              string
		Emails              Emails
		Username            *string
		WebauthnCredentials ImportWebauthnCredentials
		Password            *ImportPasswordCredential
		OTPSecret           *ImportOTPSecret
		CreatedAt           *time.Time
		UpdatedAt           *time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "User with one primary email must validate",
			fields: fields{
				UserID: "",
				Emails: Emails{
					ImportOrExportEmail{
						Address:    "primary@hanko.io",
						IsPrimary:  true,
						IsVerified: false,
					},
				},
				CreatedAt: nil,
				UpdatedAt: nil,
			},
			wantErr: assert.NoError,
		},
		{
			name: "User with no email but username must validate",
			fields: fields{
				Username: &validUsername,
			},
			wantErr: assert.NoError,
		},
		{
			name: "UserID with valid uuid must validate",
			fields: fields{
				UserID: validUUID,
				Emails: Emails{
					ImportOrExportEmail{
						Address:    "primary@hanko.io",
						IsPrimary:  true,
						IsVerified: false,
					},
				},
				CreatedAt: nil,
				UpdatedAt: nil,
			},
			wantErr: assert.NoError,
		},
		{
			name: "UserID with invalid uuid must not validate",
			fields: fields{
				UserID: invalidUUID,
				Emails: Emails{
					ImportOrExportEmail{
						Address:    "primary@hanko.io",
						IsPrimary:  true,
						IsVerified: false,
					},
				},
				CreatedAt: nil,
				UpdatedAt: nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "User with no email and username must not validate",
			fields: fields{
				UserID:    "",
				Emails:    nil,
				CreatedAt: nil,
				UpdatedAt: nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "User with an empty username must not validate",
			fields: fields{
				Emails: Emails{
					ImportOrExportEmail{
						Address:    "primary@hanko.io",
						IsPrimary:  true,
						IsVerified: false,
					},
				},
				Username: &emptyUsername,
			},
			wantErr: assert.Error,
		},
		{
			name: "User with no primary email must not validate",
			fields: fields{
				UserID: "",
				Emails: Emails{
					ImportOrExportEmail{
						Address:    "primary@hanko.io",
						IsPrimary:  false,
						IsVerified: false,
					},
				},
				CreatedAt: nil,
				UpdatedAt: nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "More than one Primary must not validate",
			fields: fields{
				UserID: "",
				Emails: Emails{
					ImportOrExportEmail{
						Address:    "primary@hanko.io",
						IsPrimary:  true,
						IsVerified: false,
					},
					ImportOrExportEmail{
						Address:    "primary2@hanko.io",
						IsPrimary:  true,
						IsVerified: false,
					},
				},
				CreatedAt: nil,
				UpdatedAt: nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "Valid webauthn credential must validate",
			fields: fields{
				Username: &validUsername,
				WebauthnCredentials: ImportWebauthnCredentials{
					webauthnCredential,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Webauthn credential without id must not validate",
			fields: fields{
				Username: &validUsername,
				WebauthnCredentials: ImportWebauthnCredentials{
					webauthnCredentialWithEmptyID,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Webauthn credential without public key must not validate",
			fields: fields{
				Username: &validUsername,
				WebauthnCredentials: ImportWebauthnCredentials{
					webauthnCredentialWithEmptyPublicKey,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Webauthn credential without attestation type must not validate",
			fields: fields{
				Username: &validUsername,
				WebauthnCredentials: ImportWebauthnCredentials{
					webauthnCredentialWithEmptyAttestationType,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "User with password must validate",
			fields: fields{
				Username: &validUsername,
				Password: &ImportPasswordCredential{
					Password: "$2a$12$mFbud0mLsD/q.WG7/9pNQemlAHs3H4o8zAv44gsUF1v1awsdqTh7.",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "User with empty password string must not validate",
			fields: fields{
				Username: &validUsername,
				Password: &ImportPasswordCredential{
					Password: "",
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Wrong formatted password must not validate",
			fields: fields{
				Username: &validUsername,
				Password: &ImportPasswordCredential{
					Password: "$12$mFbud0mLsD/q.WG7/9pNQemlAHs3H4o8zAv44gsUF1v1awsdqTh7.",
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "User with OTPSecret must validate",
			fields: fields{
				Username: &validUsername,
				OTPSecret: &ImportOTPSecret{
					Secret: "MYOTPSECRET",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "User with empty OTPSecret string must not validate",
			fields: fields{
				Username: &validUsername,
				OTPSecret: &ImportOTPSecret{
					Secret: "",
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &ImportOrExportEntry{
				UserID:              tt.fields.UserID,
				Emails:              tt.fields.Emails,
				CreatedAt:           tt.fields.CreatedAt,
				UpdatedAt:           tt.fields.UpdatedAt,
				Username:            tt.fields.Username,
				WebauthnCredentials: tt.fields.WebauthnCredentials,
				Password:            tt.fields.Password,
				OTPSecret:           tt.fields.OTPSecret,
			}
			tt.wantErr(t, entry.validate(v), fmt.Sprintf("validate()"))
		})
	}
}

var webauthnCredential = ImportWebauthnCredential{
	ID:              "randomID",
	Name:            nil,
	PublicKey:       "randomPublicKey",
	AttestationType: "none",
	AAGUID:          uuid.Nil,
	SignCount:       0,
	LastUsedAt:      nil,
	CreatedAt:       nil,
	UpdatedAt:       nil,
	Transports:      nil,
	BackupEligible:  false,
	BackupState:     false,
	MFAOnly:         false,
}

var webauthnCredentialWithEmptyID = ImportWebauthnCredential{
	ID:              "",
	Name:            nil,
	PublicKey:       "randomPublicKey",
	AttestationType: "none",
	AAGUID:          uuid.Nil,
	SignCount:       0,
	LastUsedAt:      nil,
	CreatedAt:       nil,
	UpdatedAt:       nil,
	Transports:      nil,
	BackupEligible:  false,
	BackupState:     false,
	MFAOnly:         false,
}

var webauthnCredentialWithEmptyPublicKey = ImportWebauthnCredential{
	ID:              "randomID",
	Name:            nil,
	PublicKey:       "",
	AttestationType: "none",
	AAGUID:          uuid.Nil,
	SignCount:       0,
	LastUsedAt:      nil,
	CreatedAt:       nil,
	UpdatedAt:       nil,
	Transports:      nil,
	BackupEligible:  false,
	BackupState:     false,
	MFAOnly:         false,
}

var webauthnCredentialWithEmptyAttestationType = ImportWebauthnCredential{
	ID:              "randomID",
	Name:            nil,
	PublicKey:       "randomPublicKey",
	AttestationType: "",
	AAGUID:          uuid.Nil,
	SignCount:       0,
	LastUsedAt:      nil,
	CreatedAt:       nil,
	UpdatedAt:       nil,
	Transports:      nil,
	BackupEligible:  false,
	BackupState:     false,
	MFAOnly:         false,
}
