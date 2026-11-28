package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A user without an email address has nobody to notify, so SendNotification must
// return early. The zero-value service and params would panic if it went on.
func TestSendNotification_NoEmailAddressIsSkipped(t *testing.T) {
	var s securityNotification

	err := s.SendNotification(nil, SendSecurityNotificationParams{Template: "mfa_create"})

	assert.NoError(t, err)
}
