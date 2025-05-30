package handler

import (
	"errors"
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/rate_limiter"
	"github.com/teamhanko/hanko/backend/session"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHandler struct {
	persister      persistence.Persister
	sessionManager session.Manager
	cfg            *config.Config
	auditLogger    auditlog.Logger
	rateLimiter    limiter.Store
}

func NewPasswordHandler(persister persistence.Persister, sessionManager session.Manager, cfg *config.Config, auditLogger auditlog.Logger) *PasswordHandler {
	var rateLimiter limiter.Store
	if cfg.RateLimiter.Enabled {
		rateLimiter = rate_limiter.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.PasswordLimits)
	}
	return &PasswordHandler{
		persister:      persister,
		sessionManager: sessionManager,
		cfg:            cfg,
		auditLogger:    auditLogger,
		rateLimiter:    rateLimiter,
	}
}

type PasswordSetBody struct {
	UserID   string `json:"user_id" validate:"required,uuid"`
	Password string `json:"password" validate:"required"`
}

func (h *PasswordHandler) Set(c echo.Context) error {
	var body PasswordSetBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("missing or malformed jwt")
	}

	sessionUserId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	user, err := h.persister.GetUserPersister().Get(uuid.FromStringOrNil(body.UserID))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	pwBytes := []byte(body.Password)
	if utf8.RuneCountInString(body.Password) < h.cfg.Password.MinLength { // use utf8.RuneCountInString, so utf8 characters would count as 1
		err = h.auditLogger.Create(c, models.AuditLogPasswordSetFailed, user, fmt.Errorf("password too short"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("password must be at least %d characters long", h.cfg.Password.MinLength))
	}

	if len(pwBytes) > 72 {
		err = h.auditLogger.Create(c, models.AuditLogPasswordSetFailed, user, fmt.Errorf("password too long"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "password must not be longer than 72 bytes")
	}

	if user == nil {
		err = h.auditLogger.Create(c, models.AuditLogPasswordSetFailed, user, fmt.Errorf("unknown user: %s", body.UserID))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New(fmt.Sprintf("user %s not found ", sessionUserId)))
	}

	if sessionUserId != user.ID {
		err = h.auditLogger.Create(c, models.AuditLogPasswordSetFailed, user, fmt.Errorf("wrong user: expected %s -> got %s", sessionUserId, user.ID))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusForbidden).SetInternal(errors.New(fmt.Sprintf("session.userId %s tried to set password credentials for body.userId %s", sessionUserId, user.ID)))
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		pwPersister := h.persister.GetPasswordCredentialPersisterWithConnection(tx)
		pw, err := pwPersister.GetByUserID(user.ID)
		if err != nil {
			return fmt.Errorf("failed to get credential: %w", err)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(pwBytes, 12)
		if err != nil {
			return fmt.Errorf("failed to hash password: %s", err)
		}

		newPw := models.PasswordCredential{
			UserId:   uuid.FromStringOrNil(body.UserID),
			Password: string(hashedPassword),
		}

		if pw == nil {
			err = pwPersister.Create(newPw)
			if err != nil {
				return fmt.Errorf("failed to create password: %w", err)
			} else {
				err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogPasswordSetSucceeded, user, nil)
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
				return c.JSON(http.StatusCreated, nil)
			}
		} else {
			newPw.ID = pw.ID
			err = pwPersister.Update(newPw)
			if err != nil {
				return fmt.Errorf("failed to set password: %w", err)
			} else {
				err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogPasswordSetSucceeded, user, nil)
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
				return c.JSON(http.StatusOK, nil)
			}
		}
	})
}

type PasswordLoginBody struct {
	UserId   string `json:"user_id" validate:"required,uuid"`
	Password string `json:"password" validate:"required"`
}

func (h *PasswordHandler) Login(c echo.Context) error {
	var body PasswordLoginBody
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	userId, err := uuid.FromString(body.UserId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is not a uuid").SetInternal(err)
	}

	if h.rateLimiter != nil {
		err := rate_limiter.Limit(h.rateLimiter, userId, c)
		if err != nil {
			return err
		}
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		err = h.auditLogger.Create(c, models.AuditLogPasswordLoginFailed, nil, fmt.Errorf("unknown user: %s", userId))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("user not found"))
	}

	pwBytes := []byte(body.Password)
	if len(pwBytes) > 72 {
		err = h.auditLogger.Create(c, models.AuditLogPasswordLoginFailed, user, errors.New("password too long"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "password must not be longer than 72 bytes")
	}

	pw, err := h.persister.GetPasswordCredentialPersister().GetByUserID(uuid.FromStringOrNil(body.UserId))
	if pw == nil {
		err = h.auditLogger.Create(c, models.AuditLogPasswordLoginFailed, user, fmt.Errorf("user has no password credential"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New(fmt.Sprintf("no password credential found for: %s", body.UserId)))
	}

	if err != nil {
		return fmt.Errorf("error retrieving credential: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(pw.Password), pwBytes); err != nil {
		err = h.auditLogger.Create(c, models.AuditLogPasswordLoginFailed, user, fmt.Errorf("password hash not equal"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusUnauthorized).SetInternal(err)
	}

	var emailJwt *dto.EmailJWT
	if e := user.Emails.GetPrimary(); e != nil {
		emailJwt = dto.EmailJWTFromEmailModel(e)
	}

	token, rawToken, err := h.sessionManager.GenerateJWT(dto.UserJWT{
		UserID: pw.UserId.String(),
		Email:  emailJwt,
	})
	if err != nil {
		return fmt.Errorf("failed to generate jwt: %w", err)
	}

	cookie, err := h.sessionManager.GenerateCookie(token)
	if err != nil {
		return fmt.Errorf("failed to create session cookie: %w", err)
	}

	err = storeSession(h.cfg, h.persister, userId, rawToken, c, h.persister.GetConnection())
	if err != nil {
		return fmt.Errorf("failed to store session in DB: %w", err)
	}

	c.Response().Header().Set("X-Session-Lifetime", fmt.Sprintf("%d", cookie.MaxAge))

	if h.cfg.Session.EnableAuthTokenHeader {
		c.Response().Header().Set("X-Auth-Token", token)
	} else {
		c.SetCookie(cookie)
	}

	err = h.auditLogger.Create(c, models.AuditLogPasswordLoginSucceeded, user, nil)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return c.JSON(http.StatusOK, nil)
}
