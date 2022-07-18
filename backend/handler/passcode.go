package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"net/http"
	"time"
)

type PasscodeHandler struct {
	mailer            mail.Mailer
	renderer          *mail.Renderer
	passcodeGenerator crypto.PasscodeGenerator
	persister         persistence.Persister
	emailConfig       config.Email
	serviceConfig     config.Service
	TTL               int
	sessionManager    session.Manager
	cfg               *config.Config
}

var maxPasscodeTries = 3

func NewPasscodeHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, mailer mail.Mailer) (*PasscodeHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	return &PasscodeHandler{
		mailer:            mailer,
		renderer:          renderer,
		passcodeGenerator: crypto.NewPasscodeGenerator(),
		persister:         persister,
		emailConfig:       cfg.Passcode.Email,
		serviceConfig:     cfg.Service,
		TTL:               cfg.Passcode.TTL,
		sessionManager:    sessionManager,
		cfg:               cfg,
	}, nil
}

func (h *PasscodeHandler) Init(c echo.Context) error {
	var body dto.PasscodeInitRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	userId, err := uuid.FromString(body.UserId)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return dto.NewHTTPError(http.StatusBadRequest).SetInternal(errors.New("user not found"))
	}

	passcode, err := h.passcodeGenerator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate passcode: %w", err)
	}

	passcodeId, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create passcodeId: %w", err)
	}
	now := time.Now()
	hashedPasscode, err := bcrypt.GenerateFromPassword([]byte(passcode), 12)
	if err != nil {
		return fmt.Errorf("failed to hash passcode: %w", err)
	}
	passcodeModel := models.Passcode{
		ID:        passcodeId,
		UserId:    userId,
		Ttl:       h.TTL,
		Code:      string(hashedPasscode),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = h.persister.GetPasscodePersister().Create(passcodeModel)
	if err != nil {
		return fmt.Errorf("failed to store passcode: %w", err)
	}

	durationTTL := time.Duration(h.TTL) * time.Second
	data := map[string]interface{}{
		"Code":        passcode,
		"ServiceName": h.serviceConfig.Name,
		"TTL":         fmt.Sprintf("%.0f", durationTTL.Minutes()),
	}

	lang := c.Request().Header.Get("Accept-Language")
	str, err := h.renderer.Render("loginTextMail", lang, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	message := gomail.NewMessage()
	message.SetAddressHeader("To", user.Email, "")
	message.SetAddressHeader("From", h.emailConfig.FromAddress, h.emailConfig.FromName)

	message.SetHeader("Subject", h.renderer.Translate(lang, "email_subject_login", data))

	message.SetBody("text/plain", str)

	err = h.mailer.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send passcode: %w", err)
	}

	return c.JSON(http.StatusOK, dto.PasscodeReturn{
		Id:        passcodeId.String(),
		TTL:       h.TTL,
		CreatedAt: passcodeModel.CreatedAt,
	})
}

func (h *PasscodeHandler) Finish(c echo.Context) error {
	startTime := time.Now()
	var body dto.PasscodeFinishRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	passcodeId, err := uuid.FromString(body.Id)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse passcodeId as uuid").SetInternal(err)
	}

	// only if an internal server occurs the transaction should be rolled back
	var businessError error
	transactionError := h.persister.Transaction(func(tx *pop.Connection) error {
		passcodePersister := h.persister.GetPasscodePersisterWithConnection(tx)
		userPersister := h.persister.GetUserPersisterWithConnection(tx)
		passcode, err := passcodePersister.Get(passcodeId)
		if err != nil {
			return fmt.Errorf("failed to get passcode: %w", err)
		}
		if passcode == nil {
			businessError = dto.NewHTTPError(http.StatusNotFound, "passcode not found")
			return nil
		}

		lastVerificationTime := passcode.CreatedAt.Add(time.Duration(passcode.Ttl) * time.Second)
		if lastVerificationTime.Before(startTime) {
			businessError = dto.NewHTTPError(http.StatusRequestTimeout, "passcode request timed out").SetInternal(errors.New(fmt.Sprintf("createdAt: %s -> lastVerificationTime: %s", passcode.CreatedAt, lastVerificationTime))) // TODO: maybe we should use BadRequest, because RequestTimeout might be to technical and can refer to different error
			return nil
		}

		err = bcrypt.CompareHashAndPassword([]byte(passcode.Code), []byte(body.Code))
		if err != nil {
			passcode.TryCount = passcode.TryCount + 1

			if passcode.TryCount >= maxPasscodeTries {
				err = passcodePersister.Delete(*passcode)
				if err != nil {
					return fmt.Errorf("failed to delete passcode: %w", err)
				}
				businessError = dto.NewHTTPError(http.StatusGone, "max attempts reached")
				return nil
			}

			err = passcodePersister.Update(*passcode)
			if err != nil {
				return fmt.Errorf("failed to update passcode: %w", err)
			}

			businessError = dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("passcode invalid"))
			return nil
		}

		err = passcodePersister.Delete(*passcode)
		if err != nil {
			return fmt.Errorf("failed to delete passcode: %w", err)
		}

		user, err := userPersister.Get(passcode.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if !user.Verified {
			user.Verified = true
			err = userPersister.Update(*user)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
		}

		token, err := h.sessionManager.GenerateJWT(passcode.UserId)
		if err != nil {
			return fmt.Errorf("failed to generate jwt: %w", err)
		}

		cookie, err := h.sessionManager.GenerateCookie(token)
		if err != nil {
			return fmt.Errorf("failed to create session token: %w", err)
		}

		c.SetCookie(cookie)

		if h.cfg.Session.EnableAuthToken {
			c.Response().Header().Set("X-Auth-Token", token)
		}

		return c.JSON(http.StatusOK, dto.PasscodeReturn{
			Id:        passcode.ID.String(),
			TTL:       passcode.Ttl,
			CreatedAt: passcode.CreatedAt,
		})
	})

	if businessError != nil {
		return businessError
	}

	return transactionError
}
