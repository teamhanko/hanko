package flow_api_test

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"strings"
)

func initPasskey(c flowpilot.ExecutionContext) {
	passkeyChallenge := "42"
	_ = c.Stash().Set("passkey_public_key", passkeyChallenge)
	_ = c.Payload().Set("challenge", passkeyChallenge)
}

func initPasscode(c flowpilot.ExecutionContext, email string, generate2faToken bool) {
	passcodeIDStash := c.Stash().Get("passcode_id")
	emailStash := c.Stash().Get("email")

	if len(passcodeIDStash.String()) == 0 || emailStash.String() != email {
		// resend passcode
		id, _ := uuid.NewV4()
		_ = c.Stash().Set("passcode_id", id.String())
		_ = c.Input().Set("passcode_id", id.String())
	} else {
		_ = c.Input().Set("passcode_id", passcodeIDStash.String())
	}

	_ = c.Stash().Set("email", email)
	_ = c.Stash().Set("code", "424242")

	if generate2faToken {
		token, _ := generateToken(32)
		_ = c.Stash().Set("passcode_2fa_token", token)
		_ = c.Input().Set("passcode_2fa_token", token)
	}
}

func generateToken(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	randomString = strings.ReplaceAll(randomString, "-", "")
	randomString = strings.ReplaceAll(randomString, "_", "")
	return randomString[:length], nil
}
