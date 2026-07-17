package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMFA_Validate_AcceptsParseableDuration(t *testing.T) {
	mfa := MFA{DeviceTrustDuration: "720h"}
	assert.NoError(t, mfa.Validate())
}

func TestMFA_Validate_RejectsUnparseableDuration(t *testing.T) {
	mfa := MFA{DeviceTrustDuration: "not-a-duration"}
	assert.Error(t, mfa.Validate())
}

func TestMFA_DeviceTrustDurationMarshalsAsPlainString(t *testing.T) {
	mfa := MFA{DeviceTrustDuration: "720h"}

	raw, err := json.Marshal(mfa)
	require.NoError(t, err)

	var doc map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &doc))

	assert.Equal(t, "720h", doc["device_trust_duration"])
}
