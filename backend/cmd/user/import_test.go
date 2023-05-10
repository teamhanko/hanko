package user

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestValidate(t *testing.T) {
	moreThanOnePrimary := []ImportEntry{
		{
			Emails: Emails{
				{
					Address:    "1@example.com",
					IsPrimary:  true,
					IsVerified: false,
				},
				{
					Address:    "2@example.com",
					IsPrimary:  true,
					IsVerified: false,
				},
			},
		},
	}
	err := validateEntries(moreThanOnePrimary)
	log.Println(err)
	assert.Error(t, err)
	noEmails := []ImportEntry{
		{
			UserID: "someId",
			Emails: Emails{},
		},
	}
	err = validateEntries(noEmails)
	log.Println(err)
	assert.Error(t, err)
}
