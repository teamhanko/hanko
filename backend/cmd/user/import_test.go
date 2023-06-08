package user

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
	"time"
)

func Test_loadFromFile(t *testing.T) {
	type args struct {
		input io.Reader
	}
	standardTime, _ := time.Parse(time.RFC3339, "2023-06-07T13:42:49.369489Z")
	tests := []struct {
		name    string
		args    args
		want    []ImportEntry
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty array -> empty result",
			args: args{
				input: strings.NewReader("[]"),
			},
			wantErr: assert.NoError,
			want:    []ImportEntry{},
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
			want: []ImportEntry{
				{
					UserID: "799e95f0-4cc7-4bd7-9f01-5fdc4fa26ea3",
					Emails: Emails{
						ImportEmail{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadFromFile(tt.args.input)
			if !tt.wantErr(t, err, fmt.Sprintf("loadFromFile(%v)", tt.args.input)) {
				return
			}
			assert.Equalf(t, tt.want, got, "loadFromFile(%v)", tt.args.input)
		})
	}
}
