package models

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEmails_GetPrimaryAddress(t *testing.T) {
	t.Run("no emails", func(t *testing.T) {
		emails := Emails{}
		assert.Equal(t, "", emails.GetPrimaryAddress())
	})

	t.Run("emails but none primary", func(t *testing.T) {
		emails := Emails{{Address: "a@example.com"}}
		assert.Equal(t, "", emails.GetPrimaryAddress())
	})

	t.Run("primary email", func(t *testing.T) {
		id, err := uuid.NewV4()
		assert.NoError(t, err)
		emails := Emails{
			{Address: "a@example.com"},
			{Address: "b@example.com", PrimaryEmail: &PrimaryEmail{ID: id}},
		}
		assert.Equal(t, "b@example.com", emails.GetPrimaryAddress())
	})
}
