package handler

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	"github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/rate_limiter"
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
	auditLogger       auditlog.Logger
	rateLimiter       limiter.Store
}

var maxPasscodeTries = 3

func NewPasscodeHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, mailer mail.Mailer, auditLogger auditlog.Logger) (*PasscodeHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	var rateLimiter limiter.Store
	if cfg.RateLimiter.Enabled {
		rateLimiter = rate_limiter.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.PasscodeLimits)
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
		auditLogger:       auditLogger,
		rateLimiter:       rateLimiter,
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
		err = h.auditLogger.Create(c, models.AuditLogPasscodeLoginInitFailed, nil, fmt.Errorf("unknown user"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return dto.NewHTTPError(http.StatusBadRequest).SetInternal(errors.New("user not found"))
	}

	if h.rateLimiter != nil {
		err := rate_limiter.Limit(h.rateLimiter, userId, c)
		if err != nil {
			return err
		}
	}

	var emailId uuid.UUID
	if body.EmailId != nil {
		emailId, err = uuid.FromString(*body.EmailId)
		if err != nil {
			return dto.NewHTTPError(http.StatusBadRequest, "failed to parse emailId as uuid").SetInternal(err)
		}
	}

	// Determine where to send the passcode
	var email *models.Email
	if !emailId.IsNil() {
		// Send the passcode to the specified email address
		email, err = h.persister.GetEmailPersister().Get(emailId)
		if email == nil {
			return dto.NewHTTPError(http.StatusBadRequest, "the specified emailId is not available")
		}
	} else if e := user.Emails.GetPrimary(); e == nil {
		// Workaround to support hanko element versions before v0.1.0-alpha:
		// If user has no primary email, check if a cookie with an email id is present
		emailIdCookie, err := c.Cookie("hanko_email_id")
		if err != nil {
			return fmt.Errorf("failed to get email id cookie: %w", err)
		}

		if emailIdCookie != nil && emailIdCookie.Value != "" {
			emailId, err = uuid.FromString(emailIdCookie.Value)
			if err != nil {
				return dto.NewHTTPError(http.StatusBadRequest, "failed to parse emailId as uuid").SetInternal(err)
			}
			email, err = h.persister.GetEmailPersister().Get(emailId)
			if email == nil {
				return dto.NewHTTPError(http.StatusBadRequest, "the specified emailId is not available")
			}
		} else {
			// Can't determine email address to which the passcode should be sent to
			return dto.NewHTTPError(http.StatusBadRequest, "an emailId needs to be specified")
		}
	} else {
		// Send the passcode to the primary email address
		email = e
	}

	if email.User != nil && email.User.ID.String() != user.ID.String() {
		return dto.NewHTTPError(http.StatusForbidden).SetInternal(errors.New("email address is assigned to another user"))
	}

	passcode, err := h.passcodeGenerator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate passcode: %w", err)
	}

	passcodeId, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create passcodeId: %w", err)
	}
	now := time.Now().UTC()
	hashedPasscode, err := bcrypt.GenerateFromPassword([]byte(passcode), 12)
	if err != nil {
		return fmt.Errorf("failed to hash passcode: %w", err)
	}
	passcodeModel := models.Passcode{
		ID:        passcodeId,
		UserId:    userId,
		EmailID:   email.ID,
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
	message.SetAddressHeader("To", email.Address, "")
	message.SetAddressHeader("From", h.emailConfig.FromAddress, h.emailConfig.FromName)

	message.SetHeader("Subject", h.renderer.Translate(lang, "email_subject_login", data))

	message.SetBody("text/plain", str)

	err = h.mailer.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send passcode: %w", err)
	}

	err = h.auditLogger.Create(c, models.AuditLogPasscodeLoginInitSucceeded, user, nil)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return c.JSON(http.StatusOK, dto.PasscodeReturn{
		Id:        passcodeId.String(),
		TTL:       h.TTL,
		CreatedAt: passcodeModel.CreatedAt,
	})
}

func (h *PasscodeHandler) Finish(c echo.Context) error {
	startTime := time.Now().UTC()
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

	// only if an internal server error occurs the transaction should be rolled back
	var businessError error
	transactionError := h.persister.Transaction(func(tx *pop.Connection) error {
		passcodePersister := h.persister.GetPasscodePersisterWithConnection(tx)
		userPersister := h.persister.GetUserPersisterWithConnection(tx)
		emailPersister := h.persister.GetEmailPersisterWithConnection(tx)
		primaryEmailPersister := h.persister.GetPrimaryEmailPersisterWithConnection(tx)
		passcode, err := passcodePersister.Get(passcodeId)
		if err != nil {
			return fmt.Errorf("failed to get passcode: %w", err)
		}
		if passcode == nil {
			err = h.auditLogger.Create(c, models.AuditLogPasscodeLoginFinalFailed, nil, fmt.Errorf("unknown passcode"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			businessError = dto.NewHTTPError(http.StatusUnauthorized, "passcode not found")
			return nil
		}

		user, err := userPersister.Get(passcode.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		lastVerificationTime := passcode.CreatedAt.Add(time.Duration(passcode.Ttl) * time.Second)
		if lastVerificationTime.Before(startTime) {
			err = h.auditLogger.Create(c, models.AuditLogPasscodeLoginFinalFailed, user, fmt.Errorf("timed out passcode"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
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
				err = h.auditLogger.Create(c, models.AuditLogPasscodeLoginFinalFailed, user, fmt.Errorf("max attempts reached"))
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
				businessError = dto.NewHTTPError(http.StatusGone, "max attempts reached")
				return nil
			}

			err = passcodePersister.Update(*passcode)
			if err != nil {
				return fmt.Errorf("failed to update passcode: %w", err)
			}

			err = h.auditLogger.Create(c, models.AuditLogPasscodeLoginFinalFailed, user, fmt.Errorf("passcode invalid"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			businessError = dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("passcode invalid"))
			return nil
		}

		err = passcodePersister.Delete(*passcode)
		if err != nil {
			return fmt.Errorf("failed to delete passcode: %w", err)
		}

		if passcode.Email.User != nil && passcode.Email.User.ID.String() != user.ID.String() {
			return dto.NewHTTPError(http.StatusForbidden, "email address has been claimed by another user")
		}

		if !passcode.Email.Verified {
			// Update email verified status and assign the email address to the user.
			passcode.Email.Verified = true
			passcode.Email.UserID = &user.ID

			err = emailPersister.Update(passcode.Email)
			if err != nil {
				return fmt.Errorf("failed to update the email verified status: %w", err)
			}

			if user.Emails.GetPrimary() == nil {
				primaryEmail := models.NewPrimaryEmail(passcode.Email.ID, user.ID)
				err = primaryEmailPersister.Create(*primaryEmail)
				if err != nil {
					return fmt.Errorf("failed to create primary email: %w", err)
				}

				user.Emails = models.Emails{passcode.Email}
				user.Emails.SetPrimary(primaryEmail)
				err = h.auditLogger.Create(c, models.AuditLogPrimaryEmailChanged, user, nil)
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
			}

			err = h.auditLogger.Create(c, models.AuditLogEmailVerified, user, nil)
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
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

		if h.cfg.Session.EnableAuthTokenHeader {
			c.Response().Header().Set("X-Auth-Token", token)
			//c.Response().Header().Set("Access-Control-Expose-Headers", "X-Auth-Token")
		}

		err = h.auditLogger.Create(c, models.AuditLogPasscodeLoginFinalSucceeded, user, nil)
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
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
