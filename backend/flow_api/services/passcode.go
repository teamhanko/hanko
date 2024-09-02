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

type SendPasscodeParams struct {
	Template     string
	EmailAddress string
	Language     string
}

type ValidatePasscodeParams struct {
	Tx         *pop.Connection
	PasscodeID uuid.UUID
}

type SendPasscodeResult struct {
	PasscodeModel models.Passcode
	Subject       string
	Body          string
	Code          string
}

type Passcode interface {
	ValidatePasscode(ValidatePasscodeParams) (bool, error)
	SendPasscode(*pop.Connection, SendPasscodeParams) (*SendPasscodeResult, error)
	VerifyPasscodeCode(tx *pop.Connection, passcodeID uuid.UUID, passcode string) error
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

func (s *passcode) ValidatePasscode(p ValidatePasscodeParams) (bool, error) {
	if !p.PasscodeID.IsNil() {
		_, err := s.getPasscode(p.Tx, p.PasscodeID)
		if err != nil {
			if errors.Is(err, ErrorPasscodeNotFound) || errors.Is(err, ErrorPasscodeExpired) || errors.Is(err, ErrorPasscodeMaxAttemptsReached) {
				return false, nil
			} else {
				return false, fmt.Errorf("failed to get passcode from db: %v", err)
			}
		}

		return true, nil
	}

	return false, nil
}

func (s *passcode) VerifyPasscodeCode(tx *pop.Connection, passcodeID uuid.UUID, value string) error {
	passcodePersister := s.persister.GetPasscodePersisterWithConnection(tx)
	passcodeModel, err := s.getPasscode(tx, passcodeID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passcodeModel.Code), []byte(value))
	if err != nil {
		passcodeModel.TryCount += 1

		err = passcodePersister.Update(*passcodeModel)
		if err != nil {
			return fmt.Errorf("failed to update passcode: %w", err)
		}

		if passcodeModel.TryCount >= maxPasscodeTries {
			return ErrorPasscodeMaxAttemptsReached
		}

		return ErrorPasscodeInvalid
	}

	err = passcodePersister.Delete(*passcodeModel)
	if err != nil {
		return fmt.Errorf("failed to delete passcode from db: %w", err)
	}

	return nil
}

func (s *passcode) SendPasscode(tx *pop.Connection, p SendPasscodeParams) (*SendPasscodeResult, error) {
	code, err := s.passcodeGenerator.Generate()
	if err != nil {
		return nil, err
	}
	hashedPasscode, err := bcrypt.GenerateFromPassword([]byte(code), 12)
	if err != nil {
		return nil, err
	}

	passcodeId, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	passcodeModel := models.Passcode{
		ID:        passcodeId,
		Ttl:       s.cfg.Email.PasscodeTtl,
		Code:      string(hashedPasscode),
		TryCount:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.persister.GetPasscodePersisterWithConnection(tx).Create(passcodeModel)
	if err != nil {
		return nil, err
	}

	durationTTL := time.Duration(passcodeModel.Ttl) * time.Second
	data := map[string]interface{}{
		"Code":        code,
		"ServiceName": s.cfg.Service.Name,
		"TTL":         fmt.Sprintf("%.0f", durationTTL.Minutes()),
	}

	subject := s.emailService.RenderSubject(p.Language, p.Template, data)
	body, err := s.emailService.RenderBody(p.Language, p.Template, data)
	if err != nil {
		return nil, err
	}

	if s.cfg.EmailDelivery.Enabled {
		err = s.emailService.SendEmail(p.EmailAddress, subject, body)
		if err != nil {
			return nil, err
		}
	}

	return &SendPasscodeResult{
		PasscodeModel: passcodeModel,
		Subject:       subject,
		Body:          body,
		Code:          code,
	}, nil
}

func (s *passcode) getPasscode(tx *pop.Connection, passcodeID uuid.UUID) (*models.Passcode, error) {
	passcodePersister := s.persister.GetPasscodePersisterWithConnection(tx)

	passcodeModel, err := passcodePersister.Get(passcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get passcode from db: %w", err)
	}

	if passcodeModel == nil {
		return nil, ErrorPasscodeNotFound
	}

	expirationTime := passcodeModel.CreatedAt.Add(time.Duration(passcodeModel.Ttl) * time.Second)
	if expirationTime.Before(time.Now().UTC()) {
		return nil, ErrorPasscodeExpired
	}

	if passcodeModel.TryCount >= maxPasscodeTries {
		return nil, ErrorPasscodeMaxAttemptsReached
	}

	return passcodeModel, nil
}
