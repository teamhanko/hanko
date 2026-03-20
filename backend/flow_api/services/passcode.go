package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"golang.org/x/crypto/bcrypt"
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
	Cfg          config.TenantConfig
	TenantID     *uuid.UUID
}

type ValidatePasscodeParams struct {
	Tx         *pop.Connection
	PasscodeID uuid.UUID
	TenantID   *uuid.UUID
}

type SendPasscodeResult struct {
	PasscodeModel models.Passcode
	Subject       string
	BodyPlain     string
	BodyHTML      string
	Code          string
}

type Passcode interface {
	ValidatePasscode(ValidatePasscodeParams) (bool, error)
	SendPasscode(*pop.Connection, SendPasscodeParams) (*SendPasscodeResult, error)
	VerifyPasscodeCode(tx *pop.Connection, passcodeID uuid.UUID, passcode string, tenantID *uuid.UUID) error
}

type passcode struct {
	emailService Email
	persister    persistence.Persister
}

func NewPasscodeService(emailService Email, persister persistence.Persister) Passcode {
	return &passcode{
		emailService,
		persister,
	}
}

func (s *passcode) ValidatePasscode(p ValidatePasscodeParams) (bool, error) {
	if !p.PasscodeID.IsNil() {
		_, err := s.getPasscode(p.Tx, p.PasscodeID, p.TenantID)
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

func (s *passcode) VerifyPasscodeCode(tx *pop.Connection, passcodeID uuid.UUID, value string, tenantID *uuid.UUID) error {
	passcodePersister := s.persister.GetPasscodePersisterWithConnection(tx)
	passcodeModel, err := s.getPasscode(tx, passcodeID, tenantID)
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
	passcodeGenerator := crypto.NewNumericPasscodeGenerator()
	switch p.Cfg.Email.PasscodeCharset {
	case config.PasscodeCharsetAlphanumeric:
		passcodeGenerator = crypto.NewAlphanumericPasscodeGenerator()
	}
	code, err := passcodeGenerator.Generate()
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
		Ttl:       p.Cfg.Email.PasscodeTtl,
		Code:      string(hashedPasscode),
		TenantID:  p.TenantID,
		TryCount:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.persister.GetPasscodePersisterWithConnection(tx).Create(passcodeModel)
	if err != nil {
		return nil, err
	}

	durationTTL := time.Duration(passcodeModel.Ttl) * time.Second

	subjectData := map[string]interface{}{
		"Code": code,
		"TTL":  fmt.Sprintf("%.0f", durationTTL.Minutes()),
	}

	subject := s.emailService.RenderSubject(p.Language, p.Template, subjectData)

	bodyData := map[string]interface{}{
		"Code":        code,
		"TTL":         fmt.Sprintf("%.0f", durationTTL.Minutes()),
		"ServiceName": p.Cfg.Service.Name,
		"Subject":     subject,
	}

	bodyPlain, err := s.emailService.RenderBodyPlain(p.Language, p.Template, bodyData)
	if err != nil {
		return nil, err
	}

	bodyHTML, err := s.emailService.RenderBodyHTML(p.Language, p.Template, bodyData)
	if err != nil {
		return nil, err
	}

	if p.Cfg.EmailDelivery.Enabled {
		err = s.emailService.SendEmail(p.Cfg.EmailDelivery, p.EmailAddress, subject, bodyPlain, bodyHTML)
		if err != nil {
			return nil, err
		}
	}

	return &SendPasscodeResult{
		PasscodeModel: passcodeModel,
		Subject:       subject,
		BodyPlain:     bodyPlain,
		BodyHTML:      bodyHTML,
		Code:          code,
	}, nil
}

func (s *passcode) getPasscode(tx *pop.Connection, passcodeID uuid.UUID, tenantID *uuid.UUID) (*models.Passcode, error) {
	passcodePersister := s.persister.GetPasscodePersisterWithConnection(tx)

	passcodeModel, err := passcodePersister.Get(passcodeID, tenantID)
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
