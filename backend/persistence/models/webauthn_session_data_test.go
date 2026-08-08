package models

import (
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebauthnSessionData_CredParamsRoundTrip(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4())
	userID := uuid.Must(uuid.NewV4())

	original := &webauthn.SessionData{
		Challenge:        "test-challenge",
		UserID:           userID.Bytes(),
		UserVerification: protocol.VerificationRequired,
		Expires:          time.Now().UTC().Add(time.Minute).Truncate(time.Second),
		CredParams:       webauthn.CredentialParametersDefault(),
	}

	sessionDataModel, err := NewWebauthnSessionDataFrom(original, WebauthnOperationRegistration, tenantID)
	require.NoError(t, err)

	roundTripped, err := sessionDataModel.ToSessionData()
	require.NoError(t, err)

	assert.Equal(t, original.CredParams, roundTripped.CredParams)
}

func TestWebauthnSessionData_ToSessionData_EmptyCredParams(t *testing.T) {
	sessionDataModel := &WebauthnSessionData{
		Challenge: "test-challenge",
	}

	sessionData, err := sessionDataModel.ToSessionData()
	require.NoError(t, err)
	assert.Empty(t, sessionData.CredParams)
}
