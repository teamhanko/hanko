package user

import (
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/test"
)

const validUUID2 = "799e95f0-4cc7-4bd7-9f01-5fdc4fa26ea3"

func TestImportSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(importSuite))
}

type importSuite struct {
	test.Suite
}

func (s *importSuite) Test_loadAndValidate() {
	type args struct {
		input io.Reader
	}
	standardTime, _ := time.Parse(time.RFC3339, "2023-06-07T13:42:49.369489Z")
	tests := []struct {
		name    string
		args    args
		want    []ImportOrExportEntry
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty array -> empty result",
			args: args{
				input: strings.NewReader("[]"),
			},
			wantErr: assert.NoError,
			want:    []ImportOrExportEntry{},
		},
		{
			name: "empty file -> nil result",
			args: args{
				input: strings.NewReader(""),
			},
			wantErr: assert.Error,
			want:    nil,
		},
		{
			name: "one user file",
			args: args{
				input: strings.NewReader("[{\"user_id\":\"799e95f0-4cc7-4bd7-9f01-5fdc4fa26ea3\",\"emails\":[{\"address\":\"koreyrath@wolff.name\",\"is_primary\":true,\"is_verified\":true}],\"created_at\":\"2023-06-07T13:42:49.369489Z\",\"updated_at\":\"2023-06-07T13:42:49.369489Z\"}]\n"),
			},
			wantErr: assert.NoError,
			want: []ImportOrExportEntry{
				{
					UserID: validUUID2,
					Emails: Emails{
						ImportOrExportEmail{
							Address:    "koreyrath@wolff.name",
							IsPrimary:  true,
							IsVerified: true,
						},
					},
					CreatedAt: &standardTime,
					UpdatedAt: &standardTime,
				},
			},
		},
		{
			name: "corrupted json input",
			args: args{
				input: strings.NewReader("[{user_id:\"799e95f0-4cc7-4bd7-9f01-5fdc4fa26ea3\",}]\n"),
			},
			wantErr: assert.Error,
			want:    nil,
		},
		{
			name: "several validation errors",
			args: args{
				input: strings.NewReader("[{\"user_id\":\"799e95f0-4cc7-4bd7-9f1-5fdc4fa26ea3\",\"emails\":[{\"address\":\"koreyrath@wolff.name\",\"is_primary\":true,\"is_verified\":true}],\"created_at\":\"2023-06-07T13:42:49.369489Z\",\"updated_at\":\"2023-06-07T13:42:49.369489Z\"},{\"user_id\":\"799e95f0-4cc7-4bd7-9f1-5fdc4fa26ea3\",\"emails\":[{\"address\":\"koreyrath@wolff.name\",\"is_primary\":false,\"is_verified\":true}],\"created_at\":\"2023-06-07T13:42:49.369489Z\",\"updated_at\":\"2023-06-07T13:42:49.369489Z\"}]\n"),
			},
			wantErr: assert.Error,
			want:    nil,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := loadAndValidate(tt.args.input)
			if !tt.wantErr(s.T(), err, fmt.Sprintf("loadAndValidate(%v)", tt.args.input)) {
				return
			}
			assert.Equalf(s.T(), tt.want, got, "loadAndValidate(%v)", tt.args.input)
		})
	}
}

func (s *importSuite) Test_addToDatabase() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	type args struct {
		entries   []ImportOrExportEntry
		persister persistence.Persister
	}
	tests := []struct {
		name         string
		args         args
		wantErr      assert.ErrorAssertionFunc
		wantNumUsers int
	}{
		{
			name: "Positive",
			args: args{
				entries: []ImportOrExportEntry{
					{
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
				},
				persister: s.Storage,
			},
			wantErr:      assert.NoError,
			wantNumUsers: 1,
		},
		{
			name: "Double uuid",
			args: args{
				entries: []ImportOrExportEntry{
					{
						UserID: validUUID,
						Emails: Emails{
							ImportOrExportEmail{
								Address:    "primary1@hanko.io",
								IsPrimary:  true,
								IsVerified: false,
							},
						},
						CreatedAt: nil,
						UpdatedAt: nil,
					},
					{
						UserID: validUUID,
						Emails: Emails{
							ImportOrExportEmail{
								Address:    "primary2@hanko.io",
								IsPrimary:  true,
								IsVerified: false,
							},
						},
						CreatedAt: nil,
						UpdatedAt: nil,
					},
				},
				persister: s.Storage,
			},
			wantErr:      assert.Error,
			wantNumUsers: 0,
		},
		{
			name: "Double primary email",
			args: args{
				entries: []ImportOrExportEntry{
					{
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
					{
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
				},
				persister: s.Storage,
			},
			wantErr:      assert.Error,
			wantNumUsers: 0,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {

			s.SetupTest()
			tt.wantErr(t, addToDatabase(tt.args.entries, tt.args.persister), fmt.Sprintf("addToDatabase(%v, %v)", tt.args.entries, tt.args.persister))
			users, err := tt.args.persister.GetUserPersister().List(0, 100, uuid.Nil, "", "")
			log.Println(users)
			s.NoError(err)
			s.Equal(tt.wantNumUsers, len(users))

			s.TearDownTest()
		})
	}
}
