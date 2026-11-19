package persistence_test

import (
	"testing"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/flow_api/services"
	"github.com/teamhanko/hanko/backend/v3/test"
)

const (
	otpFixtureSecretID = "f28b15df-6e09-4ac0-b49f-e4e2d274f939"
	otpFixtureSecret   = "RANDOMOTPSECRET"

	// otpSecondTransactionHeadStart is how long the first transaction stays open after the second one has been
	// started. A second transaction that is not held back by the lock reads the secret in that time.
	otpSecondTransactionHeadStart = 500 * time.Millisecond
)

func TestOTPSecretPersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(otpSecretPersisterSuite))
}

type otpSecretPersisterSuite struct {
	test.Suite
}

// GetForUpdate refuses to run outside a transaction.
func (s *otpSecretPersisterSuite) TestGetForUpdate_RequiresTransaction() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	_, err := s.Storage.GetOTPSecretPersister().GetForUpdate(
		uuid.FromStringOrNil(otpFixtureSecretID),
		uuid.FromStringOrNil(config.DefaultTenantID),
	)
	s.Require().ErrorContains(err, "must be called inside a transaction")
}

// Of two concurrent transactions that validate the same TOTP code for the same secret, only the first accepts
// it, and the second is rejected as already used.
func (s *otpSecretPersisterSuite) TestValidateCode_ConcurrentSameCodeAcceptedOnce() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Require().NoError(s.LoadFixtures("../test/fixtures/otp"))

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	secretID := uuid.FromStringOrNil(otpFixtureSecretID)
	totpService := services.NewTOTPService(services.DefaultTOTPOptions())

	now := time.Now().UTC()
	code, err := totp.GenerateCodeCustom(otpFixtureSecret, now, totp.ValidateOpts{
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	s.Require().NoError(err)

	// Same steps as the OTP code validation of the login flow: lock the secret, validate the code against it,
	// store the matched step.
	validateAndStore := func(tx *pop.Connection) error {
		otpSecretPersister := s.Storage.GetOTPSecretPersisterWithConnection(tx)

		secret, err := otpSecretPersister.GetForUpdate(secretID, tenantID)
		if err != nil {
			return err
		}

		matchedStep, err := totpService.ValidateCode(code, secret, now)
		if err != nil {
			return err
		}

		secret.LastValidatedStep = nulls.NewInt64(matchedStep)
		return otpSecretPersister.Update(secret)
	}

	secondDone := make(chan error, 1)

	firstErr := s.Storage.Transaction(func(tx *pop.Connection) error {
		if err := validateAndStore(tx); err != nil {
			return err
		}

		go func() {
			secondDone <- s.Storage.Transaction(validateAndStore)
		}()

		time.Sleep(otpSecondTransactionHeadStart)

		return nil
	})
	s.Require().NoError(firstErr)

	s.Require().ErrorIs(<-secondDone, services.ErrTOTPCodeAlreadyUsed)
}
