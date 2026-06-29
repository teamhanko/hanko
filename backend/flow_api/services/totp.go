package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// ErrTOTPCodeInvalid is returned by TOTP.ValidateCode when the code does not
// match the secret within the acceptance window.
var ErrTOTPCodeInvalid = errors.New("totp code invalid")

// ErrTOTPCodeAlreadyUsed is returned by TOTP.ValidateCode when the code's time
// step has already been consumed (RFC 6238 §5.2 replay protection).
var ErrTOTPCodeAlreadyUsed = errors.New("totp code already used")

// TOTPOptions configures the TOTP service. Zero values are replaced with
// defaults in NewTOTPService, matching the totp.Validate baseline.
type TOTPOptions struct {
	Period    uint          // seconds per step; default 30
	Skew      uint          // accepted steps either side of now; default 1
	Digits    otp.Digits    // code length; default otp.DigitsSix
	Algorithm otp.Algorithm // HMAC algorithm; default otp.AlgorithmSHA1
}

// DefaultTOTPOptions returns options compatible with Google Authenticator and
// most TOTP clients (RFC 6238 defaults).
func DefaultTOTPOptions() TOTPOptions {
	return TOTPOptions{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
}

// TOTP validates time-based one-time codes.
type TOTP interface {
	// ValidateCode checks code against otpSecret at time t, including RFC 6238
	// §5.2 replay detection. Returns the matched step on success so the caller
	// can persist it. Returns ErrTOTPCodeInvalid or ErrTOTPCodeAlreadyUsed on
	// rejection. When otpSecret.LastValidatedStep is NULL (initial setup), the
	// replay check is skipped.
	ValidateCode(code string, otpSecret *models.OTPSecret, t time.Time) (int64, error)
}

type totpService struct {
	opts TOTPOptions
}

// NewTOTPService creates a TOTP service. Zero fields in opts are replaced with
// DefaultTOTPOptions values. Note: Skew=0 is a valid strict mode (no drift
// tolerance) and is not overridden; use DefaultTOTPOptions() to get Skew=1.
func NewTOTPService(opts TOTPOptions) TOTP {
	defaults := DefaultTOTPOptions()
	if opts.Period == 0 {
		opts.Period = defaults.Period
	}
	if opts.Digits == 0 {
		opts.Digits = defaults.Digits
	}
	if opts.Algorithm == 0 {
		opts.Algorithm = defaults.Algorithm
	}
	return &totpService{opts: opts}
}

func (s *totpService) ValidateCode(code string, otpSecret *models.OTPSecret, t time.Time) (int64, error) {
	step, ok, err := s.validate(code, otpSecret.Secret, t)
	if err != nil {
		return 0, fmt.Errorf("totp: crypto validation failed: %w", err)
	}
	if !ok {
		return 0, ErrTOTPCodeInvalid
	}
	if otpSecret.LastValidatedStep.Valid && step <= otpSecret.LastValidatedStep.Int64 {
		return 0, ErrTOTPCodeAlreadyUsed
	}
	return step, nil
}

func (s *totpService) validate(code, secret string, t time.Time) (int64, bool, error) {
	currentStep := int64(math.Floor(float64(t.UTC().Unix()) / float64(s.opts.Period)))

	// Mirror the counter order used by the pquerna library: T, T+1, T-1, ...
	stepsToCheck := []int64{currentStep}
	for i := 1; i <= int(s.opts.Skew); i++ {
		stepsToCheck = append(stepsToCheck, currentStep+int64(i), currentStep-int64(i))
	}

	for _, step := range stepsToCheck {
		if step < 0 {
			continue
		}
		valid, err := hotp.ValidateCustom(code, uint64(step), secret, hotp.ValidateOpts{
			Digits:    s.opts.Digits,
			Algorithm: s.opts.Algorithm,
		})
		if err != nil {
			return 0, false, fmt.Errorf("totp: hotp validation failed: %w", err)
		}
		if valid {
			return step, true, nil
		}
	}
	return 0, false, nil
}
