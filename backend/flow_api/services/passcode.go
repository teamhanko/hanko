package services

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var maxPasscodeTries = 3

var (
	ErrorPasscodeInvalid            = errors.New("passcode invalid")
	ErrorPasscodeNotFound           = errors.New("passcode not found")
	ErrorPasscodeExpired            = errors.New("passcode is expired")
	ErrorPasscodeMaxAttemptsReached = errors.New("the passcode was entered wrong too many times")
)

type Passcode interface {
	SendEmailVerification(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error)
	SendLogin(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error)
	PasswordRecovery(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error)
	SendPasscode(flowID uuid.UUID, template string, emailAddress string, lang string) (uuid.UUID, error)
	VerifyPasscode(tx *pop.Connection, passcodeID uuid.UUID, passcode string) error
}

type passcode struct {
	emailService      Email
	passcodeGenerator crypto.PasscodeGenerator
	persister         persistence.Persister
	cfg               config.Config
}

func NewPasscodeService(cfg config.Config, emailService Email, persister persistence.Persister) Passcode {
	return &passcode{
		emailService,
		crypto.NewPasscodeGenerator(),
		persister,
		cfg,
	}
}

func (s *passcode) VerifyPasscode(tx *pop.Connection, passcodeId uuid.UUID, value string) error {
	passcodePersister := s.persister.GetPasscodePersisterWithConnection(tx)

	passcodeModel, err := passcodePersister.Get(passcodeId)
	if err != nil {
		return fmt.Errorf("failed to get passcode from db: %w", err)
	}

	if passcodeModel == nil {
		return ErrorPasscodeNotFound
	}

	expirationTime := passcodeModel.CreatedAt.Add(time.Duration(passcodeModel.Ttl) * time.Second)
	if expirationTime.Before(time.Now().UTC()) {
		return ErrorPasscodeExpired
	}

	err = bcrypt.CompareHashAndPassword([]byte(passcodeModel.Code), []byte(value))
	if err != nil {
		passcodeModel.TryCount += 1
		if passcodeModel.TryCount >= maxPasscodeTries {
			err = passcodePersister.Delete(*passcodeModel)
			if err != nil {
				return fmt.Errorf("failed to delete passcode from db: %w", err)
			}

			return ErrorPasscodeMaxAttemptsReached
		}

		err = passcodePersister.Update(*passcodeModel)
		if err != nil {
			return fmt.Errorf("failed to update passcode: %w", err)
		}

		return ErrorPasscodeInvalid
	}

	err = passcodePersister.Delete(*passcodeModel)
	if err != nil {
		return fmt.Errorf("failed to delete passcode from db: %w", err)
	}

	return nil
}

func (s *passcode) SendEmailVerification(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error) {
	return s.SendPasscode(flowID, "email_verification", emailAddress, lang)
}

func (s *passcode) SendLogin(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error) {
	return s.SendPasscode(flowID, "login", emailAddress, lang)
}

func (s *passcode) PasswordRecovery(flowID uuid.UUID, emailAddress string, lang string) (uuid.UUID, error) {
	return s.SendPasscode(flowID, "password_recovery", emailAddress, lang)
}

func (s *passcode) SendPasscode(flowID uuid.UUID, template string, emailAddress string, lang string) (uuid.UUID, error) {
	code, err := s.passcodeGenerator.Generate()
	if err != nil {
		return uuid.Nil, err
	}
	hashedPasscode, err := bcrypt.GenerateFromPassword([]byte(code), 12)
	if err != nil {
		return uuid.Nil, err
	}

	passcodeId, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, err
	}

	now := time.Now().UTC()
	passcodeModel := models.Passcode{
		ID:        passcodeId,
		FlowID:    &flowID,
		Ttl:       s.cfg.Passcode.TTL,
		Code:      string(hashedPasscode),
		TryCount:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.persister.GetPasscodePersister().Create(passcodeModel)
	if err != nil {
		return uuid.Nil, err
	}

	durationTTL := time.Duration(passcodeModel.Ttl) * time.Second
	data := map[string]interface{}{
		"Code":        code,
		"ServiceName": s.cfg.Service.Name,
		"TTL":         fmt.Sprintf("%.0f", durationTTL.Minutes()),
	}

	err = s.emailService.SendEmail(template, lang, data, emailAddress)
	if err != nil {
		return uuid.Nil, err
	}

	return passcodeId, nil
}
