package mail

import (
	"github.com/stretchr/testify/assert"
	"github.com/teamhanko/hanko/backend/config"
	"testing"
)

func TestNewMailer(t *testing.T) {
	tests := []struct {
		Name      string
		Input     config.SMTP
		WantError bool
	}{
		{
			Name: "create mailer successful",
			Input: config.SMTP{
				Host:     "mail.example.com",
				Port:     "123",
				User:     "example",
				Password: "example",
			},
			WantError: false,
		},
		{
			Name: "create mailer with incompatible port",
			Input: config.SMTP{
				Host:     "mail.example.com",
				Port:     "abc",
				User:     "example",
				Password: "example",
			},
			WantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mailer, err := NewMailer(test.Input)

			if test.WantError {
				assert.Error(t, err)
				assert.Empty(t, mailer)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, mailer)
			}
		})
	}
}
