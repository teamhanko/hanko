package user

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const validUUID = "62418053-a2cd-47a8-9b61-4426380d263a"
const invalidUUID = "notvalid"

func TestImportEntry_validate(t *testing.T) {
	type fields struct {
		UserID    string
		Emails    Emails
		CreatedAt *time.Time
		UpdatedAt *time.Time
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
			name: "User with no email must not validate",
			fields: fields{
				UserID:    "",
				Emails:    nil,
				CreatedAt: nil,
				UpdatedAt: nil,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &ImportOrExportEntry{
				UserID:    tt.fields.UserID,
				Emails:    tt.fields.Emails,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
			}
			tt.wantErr(t, entry.validate(), fmt.Sprintf("validate()"))
		})
	}
}
