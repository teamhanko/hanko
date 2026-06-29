package services

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

var svc = NewTOTPService(DefaultTOTPOptions())

func newTestSecret(t *testing.T) string {
	t.Helper()
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "test",
		AccountName: "user@test.com",
	})
	require.NoError(t, err)
	return key.Secret()
}

// freshSecret returns an OTPSecret with no history (LastValidatedStep NULL),
// as used during initial TOTP setup.
func freshSecret(secret string) *models.OTPSecret {
	return &models.OTPSecret{Secret: secret}
}

func TestValidateTOTPCode_CurrentStepAccepted(t *testing.T) {
	secret := newTestSecret(t)
	now := time.Now().UTC()
	code, err := totp.GenerateCode(secret, now)
	require.NoError(t, err)

	step, err := svc.ValidateCode(code, freshSecret(secret), now)
	require.NoError(t, err)
	assert.Equal(t, now.Unix()/30, step)
}

func TestValidateTOTPCode_WrongDigitsRejected(t *testing.T) {
	secret := newTestSecret(t)
	_, err := svc.ValidateCode("000000", freshSecret(secret), time.Now().UTC())
	require.ErrorIs(t, err, ErrTOTPCodeInvalid)
}

func TestValidateTOTPCode_WrongSecret(t *testing.T) {
	secret1 := newTestSecret(t)
	secret2 := newTestSecret(t)
	now := time.Now().UTC()
	code, err := totp.GenerateCode(secret1, now)
	require.NoError(t, err)

	_, err = svc.ValidateCode(code, freshSecret(secret2), now)
	require.ErrorIs(t, err, ErrTOTPCodeInvalid)
}

func TestValidateTOTPCode_PreviousStepAccepted(t *testing.T) {
	// A code generated one period in the past (T-1) must be accepted.
	secret := newTestSecret(t)
	const period = 30
	now := time.Now().UTC()
	past := now.Add(-period * time.Second)
	code, err := totp.GenerateCode(secret, past)
	require.NoError(t, err)

	step, err := svc.ValidateCode(code, freshSecret(secret), now)
	require.NoError(t, err)
	assert.Equal(t, past.Unix()/period, step, "must return the T-1 step, not T")
}

func TestValidateTOTPCode_NextStepAccepted(t *testing.T) {
	// A code generated one period in the future (T+1) must be accepted.
	secret := newTestSecret(t)
	const period = 30
	now := time.Now().UTC()
	future := now.Add(period * time.Second)
	code, err := totp.GenerateCode(secret, future)
	require.NoError(t, err)

	step, err := svc.ValidateCode(code, freshSecret(secret), now)
	require.NoError(t, err)
	assert.Equal(t, future.Unix()/period, step, "must return the T+1 step, not T")
}

func TestValidateTOTPCode_TwoPeriodsOldRejected(t *testing.T) {
	// A code from two periods ago must be rejected (outside Skew=1).
	secret := newTestSecret(t)
	const period = 30
	now := time.Now().UTC()
	twoBack := now.Add(-2 * period * time.Second)
	code, err := totp.GenerateCode(secret, twoBack)
	require.NoError(t, err)

	_, err = svc.ValidateCode(code, freshSecret(secret), now)
	require.ErrorIs(t, err, ErrTOTPCodeInvalid)
}

func TestValidateTOTPCode_FirstUseAllowed(t *testing.T) {
	// NULL LastValidatedStep (never used) must allow any valid step.
	secret := newTestSecret(t)
	now := time.Now().UTC()
	code, err := totp.GenerateCode(secret, now)
	require.NoError(t, err)

	_, err = svc.ValidateCode(code, freshSecret(secret), now)
	require.NoError(t, err)
}

func TestValidateTOTPCode_ReplayRejected(t *testing.T) {
	secret := newTestSecret(t)
	now := time.Now().UTC()
	code, err := totp.GenerateCode(secret, now)
	require.NoError(t, err)

	step := now.Unix() / 30
	otpSecret := &models.OTPSecret{
		Secret:            secret,
		LastValidatedStep: nulls.NewInt64(step), // already consumed
	}
	_, err = svc.ValidateCode(code, otpSecret, now)
	require.ErrorIs(t, err, ErrTOTPCodeAlreadyUsed)
}

func TestValidateTOTPCode_OlderSkewStepRejectedAfterAdvancement(t *testing.T) {
	// After step T is consumed, a T-1 code (within the skew window for a fresh
	// secret) must still be rejected because step T-1 <= LastValidatedStep T
	// (forward-only progression rule).
	secret := newTestSecret(t)
	const period = 30
	now := time.Now().UTC()
	tMinus1 := now.Add(-period * time.Second)

	code, err := totp.GenerateCode(secret, tMinus1)
	require.NoError(t, err)

	currentStep := now.Unix() / period
	otpSecret := &models.OTPSecret{
		Secret:            secret,
		LastValidatedStep: nulls.NewInt64(currentStep), // T already consumed
	}

	_, err = svc.ValidateCode(code, otpSecret, now)
	require.ErrorIs(t, err, ErrTOTPCodeAlreadyUsed,
		"T-1 code must be rejected once T has been consumed (forward-only progression)")
}

func TestValidateTOTPCode_NextStepAcceptedAfterPreviousValidation(t *testing.T) {
	// After step T is consumed, a code for T+1 must be accepted (T+1 > T).
	// This verifies that normal continued usage works after the first authentication.
	secret := newTestSecret(t)
	const period = 30
	now := time.Now().UTC()
	future := now.Add(period * time.Second) // T+1

	code, err := totp.GenerateCode(secret, future)
	require.NoError(t, err)

	currentStep := now.Unix() / period
	otpSecret := &models.OTPSecret{
		Secret:            secret,
		LastValidatedStep: nulls.NewInt64(currentStep), // T already consumed
	}

	step, err := svc.ValidateCode(code, otpSecret, now)
	require.NoError(t, err)
	assert.Equal(t, future.Unix()/period, step, "must return T+1 step")
}

// TestValidateTOTPCode_SetupCodeRejectedOnFirstLogin ensures that a code used
// to verify TOTP during setup cannot be replayed as the first MFA login code.
// During setup, action_otp_code_verify persists LastValidatedStep = matchedStep,
// so the persisted secret already has that step consumed when login begins.
func TestValidateTOTPCode_SetupValidatedStepCannotBeReusedForLogin(t *testing.T) {
	secret := newTestSecret(t)
	now := time.Now().UTC()
	code, err := totp.GenerateCode(secret, now)
	require.NoError(t, err)

	// Simulate what action_otp_code_verify does: validate with a fresh secret,
	// then store the returned step as LastValidatedStep before persisting.
	setupStep, err := svc.ValidateCode(code, freshSecret(secret), now)
	require.NoError(t, err)

	// Simulate the OTPSecret as it exists in the DB after setup.
	persistedSecret := &models.OTPSecret{
		Secret:            secret,
		LastValidatedStep: nulls.NewInt64(setupStep),
	}

	// The same code must be rejected when used for the first MFA login.
	_, err = svc.ValidateCode(code, persistedSecret, now)
	require.ErrorIs(t, err, ErrTOTPCodeAlreadyUsed,
		"setup verification code must not be reusable as the first MFA login code")
}
