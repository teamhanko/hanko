package handler

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/crypto"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/mail"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/persistence/models"
	"github.com/teamhanko/hanko/session"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"net/http"
	"time"
)

type passcodeInit struct {
	UserId string `json:"user_id"`
}

type passcodeReturn struct {
	Id        string    `json:"id"`
	TTL       int       `json:"ttl"`
	CreatedAt time.Time `json:"created_at"`
}

type PasscodeHandler struct {
	mailer            mail.Mailer
	renderer          *mail.Renderer
	passcodeGenerator crypto.PasscodeGenerator
	persister         persistence.Persister
	emailConfig       config.Email
	serviceConfig     config.Service
	TTL               int
	sessionManager    session.Manager
}

func NewPasscodeHandler(config config.Passcode, serviceConfig config.Service, persister persistence.Persister, sessionManager session.Manager, mailer mail.Mailer) (*PasscodeHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	return &PasscodeHandler{
		mailer:            mailer,
		renderer:          renderer,
		passcodeGenerator: crypto.NewPasscodeGenerator(),
		persister:         persister,
		emailConfig:       config.Email,
		serviceConfig:     serviceConfig,
		TTL:               config.TTL,
		sessionManager:    sessionManager,
	}, nil
}

func (h *PasscodeHandler) Init(c echo.Context) error {
	var body passcodeInit
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	userId, err := uuid.FromString(body.UserId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
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

	return c.JSON(http.StatusOK, passcodeReturn{
		Id:        passcodeId.String(),
		TTL:       h.TTL,
		CreatedAt: passcodeModel.CreatedAt,
	})
}

func (h *PasscodeHandler) Finish(c echo.Context) error {
	startTime := time.Now()
	var body passcodeFinish
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	passcodeId, err := uuid.FromString(body.Id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		passcodePersister := h.persister.GetPasscodePersisterWithConnection(tx)
		userPersister := h.persister.GetUserPersisterWithConnection(tx)
		passcode, err := passcodePersister.Get(passcodeId)
		if err != nil {
			return fmt.Errorf("failed to get passcode: %w", err)
		}
		if passcode == nil {
			return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
		}

		if passcode.CreatedAt.Add(time.Duration(passcode.Ttl) * time.Second).Before(startTime) {
			return c.JSON(http.StatusRequestTimeout, dto.NewApiError(http.StatusRequestTimeout))
		}

		err = bcrypt.CompareHashAndPassword([]byte(passcode.Code), []byte(body.Code))
		if err != nil {
			return c.JSON(http.StatusUnauthorized, dto.NewApiError(http.StatusUnauthorized))
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

		cookie, err := h.sessionManager.GenerateCookie(passcode.UserId)
		if err != nil {
			return fmt.Errorf("failed to create session token: %w", err)
		}

		c.SetCookie(cookie)
		return c.JSON(http.StatusOK, passcodeReturn{
			Id:        passcode.ID.String(),
			TTL:       passcode.Ttl,
			CreatedAt: passcode.CreatedAt,
		})
	})
}

type passcodeFinish struct {
	Id   string
	Code string
}
