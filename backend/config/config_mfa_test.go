package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMFA_MarshalJSON_DeviceTrustDurationIsAString(t *testing.T) {
	mfa := MFA{DeviceTrustDuration: 720 * time.Hour}

	raw, err := json.Marshal(mfa)
	require.NoError(t, err)

	var doc map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &doc))

	assert.Equal(t, "720h0m0s", doc["device_trust_duration"])
}

func TestMFA_UnmarshalJSON_AcceptsDurationString(t *testing.T) {
	var mfa MFA
	err := json.Unmarshal([]byte(`{"device_trust_duration": "30m"}`), &mfa)
	require.NoError(t, err)
	assert.Equal(t, 30*time.Minute, mfa.DeviceTrustDuration)
}

func TestMFA_UnmarshalJSON_StillAcceptsRawNanoseconds(t *testing.T) {
	var mfa MFA
	err := json.Unmarshal([]byte(`{"device_trust_duration": 1800000000000}`), &mfa)
	require.NoError(t, err)
	assert.Equal(t, 30*time.Minute, mfa.DeviceTrustDuration)
}

func TestMFA_MarshalJSON_RoundTrip(t *testing.T) {
	original := MFA{
		AcquireOnLogin:               true,
		DeviceTrustCookieName:        "hanko-device-token",
		DeviceTrustDuration:          720 * time.Hour,
		DeviceTrustMaxUsersPerDevice: 20,
		DeviceTrustPolicy:            "prompt",
		Enabled:                      true,
		SecurityKeys:                 SecurityKeys{Limit: 10},
		TOTP:                         TOTP{Enabled: true},
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	var roundTripped MFA
	require.NoError(t, json.Unmarshal(raw, &roundTripped))

	assert.Equal(t, original, roundTripped)
}

func TestMFA_MarshalJSON_OtherFieldsUnaffected(t *testing.T) {
	mfa := MFA{
		AcquireOnLogin:        true,
		DeviceTrustCookieName: "hanko-device-token",
		DeviceTrustDuration:   720 * time.Hour,
		Enabled:               true,
		SecurityKeys:          SecurityKeys{Enabled: true, Limit: 5},
	}

	raw, err := json.Marshal(mfa)
	require.NoError(t, err)

	var doc map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &doc))

	assert.Equal(t, true, doc["acquire_on_login"])
	assert.Equal(t, "hanko-device-token", doc["device_trust_cookie_name"])
	assert.Equal(t, true, doc["enabled"])
	securityKeys, ok := doc["security_keys"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, securityKeys["enabled"])
	assert.Equal(t, float64(5), securityKeys["limit"])
}
