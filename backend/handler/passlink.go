package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/webhook"
	"github.com/teamhanko/hanko/backend/mail"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/rate_limiter"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// TODO: garbage collect passlinks

type PasslinkHandler struct {
	mailer            mail.Mailer
	renderer          *mail.Renderer
	passlinkGenerator crypto.PasslinkGenerator
	persister         persistence.Persister
	emailConfig       config.EmailDelivery
	serviceConfig     config.Service
	URL               string
	TTL               int
	sessionManager    session.Manager
	cfg               *config.Config
	auditLogger       auditlog.Logger
	rateLimiter       limiter.Store
}

func NewPasslinkHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, mailer mail.Mailer, auditLogger auditlog.Logger) (*PasslinkHandler, error) {
	renderer, err := mail.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to create new renderer: %w", err)
	}
	var rateLimiter limiter.Store
	if cfg.RateLimiter.Enabled {
		rateLimiter = rate_limiter.NewRateLimiter(cfg.RateLimiter, cfg.RateLimiter.PasslinkLimits)
	}
	return &PasslinkHandler{
		mailer:            mailer,
		renderer:          renderer,
		passlinkGenerator: crypto.NewPasslinkGenerator(),
		persister:         persister,
		emailConfig:       cfg.EmailDelivery,
		serviceConfig:     cfg.Service,
		URL:               cfg.Passlink.URL,
		TTL:               cfg.Email.PasslinkTtl,
		sessionManager:    sessionManager,
		cfg:               cfg,
		auditLogger:       auditLogger,
		rateLimiter:       rateLimiter,
	}, nil
}

func (h *PasslinkHandler) Init(c echo.Context) error {

	var body dto.PasslinkInitRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	userId, err := uuid.FromString(body.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	user, err := h.persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		err = h.auditLogger.Create(c, models.AuditLogPasslinkLoginInitFailed, nil, fmt.Errorf("unknown user"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return echo.NewHTTPError(http.StatusBadRequest).SetInternal(errors.New("user not found"))
	}

	if h.rateLimiter != nil {
		err := rate_limiter.Limit(h.rateLimiter, userId, c)
		if err != nil {
			return err
		}
	}

	var emailId uuid.UUID
	if body.EmailID != nil {
		emailId, err = uuid.FromString(*body.EmailID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to parse emailId as uuid").SetInternal(err)
		}
	}

	// Determine where to send the passlink
	var email *models.Email
	if !emailId.IsNil() {
		// Send the passlink to the specified email address
		email, err = h.persister.GetEmailPersister().Get(emailId)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to get email by id").SetInternal(err)
		}
		if email == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "the specified emailId is not available")
		}
	} else if e := user.Emails.GetPrimary(); e != nil {
		// Send the passlink to the primary email address
		email = e
	} else {
		// Workaround to support hanko element versions before v0.1.0-alpha:
		// If user has no primary email, check if a cookie with an email id is present
		emailIdCookie, err := c.Cookie("hanko_email_id")
		if err != nil {
			return fmt.Errorf("failed to get email id cookie: %w", err)
		}

		if emailIdCookie != nil && emailIdCookie.Value != "" {
			emailId, err = uuid.FromString(emailIdCookie.Value)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to parse emailId as uuid").SetInternal(err)
			}
			email, err = h.persister.GetEmailPersister().Get(emailId)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to get email by id").SetInternal(err)
			}
			if email == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "the specified emailId is not available")
			}
		} else {
			// Can't determine email address to which the passlink should be sent to
			return echo.NewHTTPError(http.StatusBadRequest, "an emailId needs to be specified")
		}
	}

	sessionToken := h.GetSessionToken(c)
	if sessionToken != nil && sessionToken.Subject() != user.ID.String() {
		// if the user is logged in and the requested user in the body does not match the user from the session then sending and finalizing passlinks is not allowed
		return echo.NewHTTPError(http.StatusForbidden).SetInternal(errors.New("session.userId does not match requested userId"))
	}

	if email.User != nil && email.User.ID.String() != user.ID.String() {
		return echo.NewHTTPError(http.StatusForbidden).SetInternal(errors.New("email address is assigned to another user"))
	}

	redirectPath := "/"
	if strings.HasPrefix(body.RedirectPath, "/") {
		redirectPath = body.RedirectPath
	}

	now := time.Now().UTC()
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create passlinkId: %w", err)
	}
	token, err := h.passlinkGenerator.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate passlink: %w", err)
	}
	tokenHashed, err := bcrypt.GenerateFromPassword([]byte(token), 12)
	if err != nil {
		return fmt.Errorf("failed to hash passlink: %w", err)
	}

	passlinkModel := models.Passlink{
		ID:         id,
		UserId:     userId,
		EmailID:    email.ID,
		IP:         c.RealIP(),
		TTL:        h.TTL,
		LoginCount: 0,
		Reusable:   false,
		Token:      string(tokenHashed),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	redirectURL, err := h.createRedirectURL(c, id, token, redirectPath)
	if err != nil {
		return fmt.Errorf("failed to create passlink redirect URL: %w", err)
	}

	err = h.persister.GetPasslinkPersister().Create(passlinkModel)
	if err != nil {
		return fmt.Errorf("failed to store passlink: %w", err)
	}

	durationTTL := time.Duration(h.TTL) * time.Second
	data := map[string]interface{}{
		"ServiceName": h.serviceConfig.Name,
		"Token":       token,
		"URL":         redirectURL,
		"TTL":         fmt.Sprintf("%.0f", durationTTL.Minutes()),
	}

	lang := c.Request().Header.Get("Accept-Language")
	subject := h.renderer.Translate(lang, "email_subject_login_passlink", data)
	bodyPlain, err := h.renderer.Render("passlinkLoginTextMail", lang, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	webhookData := webhook.EmailSend{
		Subject:          subject,
		BodyPlain:        bodyPlain,
		ToEmailAddress:   email.Address,
		DeliveredByHanko: true,
		AcceptLanguage:   lang,
		Type:             webhook.EmailTypePasslink,
		Data: webhook.PasslinkData{
			ServiceName:  h.cfg.Service.Name,
			Token:        token,
			URL:          redirectURL,
			TTL:          h.TTL,
			ValidUntil:   passlinkModel.CreatedAt.Add(time.Duration(h.TTL) * time.Second).UTC().Unix(),
			RedirectPath: redirectPath,
			RetryLimit:   1,
		},
	}

	if h.cfg.EmailDelivery.Enabled {
		message := gomail.NewMessage()
		message.SetAddressHeader("To", email.Address, "")
		message.SetAddressHeader("From", h.emailConfig.FromAddress, h.emailConfig.FromName)

		message.SetHeader("Subject", subject)

		message.SetBody("text/plain", bodyPlain)

		err = h.mailer.Send(message)
		if err != nil {
			return fmt.Errorf("failed to send passlink: %w", err)
		}

		err = utils.TriggerWebhooks(c, events.EmailSend, webhookData)

		if err != nil {
			log.Warn().Err(err).Msg("failed to trigger webhook")
		}
	} else {
		webhookData.DeliveredByHanko = false
		err = utils.TriggerWebhooks(c, events.EmailSend, webhookData)

		if err != nil {
			return fmt.Errorf(fmt.Sprintf("failed to trigger webhook: %s", err))
		}
	}

	err = h.auditLogger.Create(c, models.AuditLogPasslinkLoginInitSucceeded, user, nil)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return c.JSON(http.StatusOK, dto.PasslinkReturn{
		ID:        id.String(),
		CreatedAt: passlinkModel.CreatedAt,
		UserID:    userId.String(),
	})
}

func (h *PasslinkHandler) Finish(c echo.Context) error {
	startTime := time.Now().UTC()
	var body dto.PasslinkFinishRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	passlinkID, err := uuid.FromString(body.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse passlinkId as uuid").SetInternal(err)
	}

	// only if an internal server error occurs the transaction should be rolled back
	var businessError error
	transactionError := h.persister.Transaction(func(tx *pop.Connection) error {
		passlinkPersister := h.persister.GetPasslinkPersisterWithConnection(tx)
		userPersister := h.persister.GetUserPersisterWithConnection(tx)
		emailPersister := h.persister.GetEmailPersisterWithConnection(tx)
		primaryEmailPersister := h.persister.GetPrimaryEmailPersisterWithConnection(tx)
		passlink, err := passlinkPersister.Get(passlinkID)
		if err != nil {
			return fmt.Errorf("failed to get passlink: %w", err)
		}
		if passlink == nil {
			err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogPasslinkLoginFinalFailed, nil, fmt.Errorf("unknown passlink"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			businessError = echo.NewHTTPError(http.StatusUnauthorized, "passlink not found")
			return nil
		}

		userModel, err := userPersister.Get(passlink.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		lastVerificationTime := passlink.CreatedAt.Add(time.Duration(passlink.TTL) * time.Second)
		if lastVerificationTime.Before(startTime) {
			err = passlinkPersister.Delete(*passlink)
			if err != nil {
				return fmt.Errorf("failed to delete passlink: %w", err)
			}

			err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogPasslinkLoginFinalFailed, userModel, fmt.Errorf("timed out passlink: createdAt: %s -> lastVerificationTime: %s", passlink.CreatedAt, lastVerificationTime))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			businessError = echo.NewHTTPError(http.StatusRequestTimeout, "passlink request timed out").SetInternal(fmt.Errorf("createdAt: %s -> lastVerificationTime: %s", passlink.CreatedAt, lastVerificationTime)) // TODO: maybe we should use BadRequest, because RequestTimeout might be too technical and can refer to different error
			return nil
		}

		err = bcrypt.CompareHashAndPassword([]byte(passlink.Token), []byte(body.Token))
		if err != nil {
			err = passlinkPersister.Delete(*passlink)
			if err != nil {
				return fmt.Errorf("failed to delete passlink: %w", err)
			}
			err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogPasslinkLoginFinalFailed, userModel, fmt.Errorf("invalid token"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			businessError = echo.NewHTTPError(http.StatusForbidden, "invalid token")
			return nil
		}

		// a passlink is valid only once, except it is explicitly marked as reusable
		// a reusable passlink token is a security risk, but might be useful to authenticate a again and again from same link (e.g. link in a newsletter)
		if passlink.Reusable {
			passlink.LoginCount += 1

			err = passlinkPersister.Update(*passlink)
			if err != nil {
				return fmt.Errorf("failed to update passlink: %w", err)
			}
		} else {
			err = passlinkPersister.Delete(*passlink)
			if err != nil {
				return fmt.Errorf("failed to delete passlink: %w", err)
			}
		}

		if passlink.Email.User != nil && passlink.Email.User.ID.String() != userModel.ID.String() {
			return echo.NewHTTPError(http.StatusForbidden, "email address has been claimed by another user")
		}

		emailExistsForUser := false
		for _, email := range userModel.Emails {
			emailExistsForUser = email.ID == passlink.Email.ID
			if emailExistsForUser {
				break
			}
		}

		existingSessionToken := h.GetSessionToken(c)
		// return forbidden when none of these cases matches
		if !((existingSessionToken == nil && emailExistsForUser) || // normal login: when user logs in and the email used is associated with the user
			(existingSessionToken == nil && len(userModel.Emails) == 0) || // register: when user register and the user has no emails
			(existingSessionToken != nil && existingSessionToken.Subject() == userModel.ID.String())) { // add email through profile: when the user adds an email while having a session and the userIds requested in the passlink and the one in the session matches
			return echo.NewHTTPError(http.StatusForbidden).SetInternal(errors.New("passlink finalization not allowed"))
		}

		wasUnverified := false
		hasEmails := len(userModel.Emails) >= 1 // check if we need to trigger a UserCreate webhook or a UserEmailCreate one

		if !passlink.Email.Verified {
			wasUnverified = true

			// Update email verified status and assign the email address to the user.
			passlink.Email.Verified = true
			passlink.Email.UserID = &userModel.ID

			err = emailPersister.Update(passlink.Email)
			if err != nil {
				return fmt.Errorf("failed to update the email verified status: %w", err)
			}

			if userModel.Emails.GetPrimary() == nil {
				primaryEmail := models.NewPrimaryEmail(passlink.Email.ID, userModel.ID)
				err = primaryEmailPersister.Create(*primaryEmail)
				if err != nil {
					return fmt.Errorf("failed to create primary email: %w", err)
				}

				userModel.Emails = models.Emails{passlink.Email}
				userModel.SetPrimaryEmail(primaryEmail)
				err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogPrimaryEmailChanged, userModel, nil)
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
			}

			err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogEmailVerified, userModel, nil)
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
		}

		var emailJwt *dto.EmailJwt
		if e := userModel.Emails.GetPrimary(); e != nil {
			emailJwt = dto.JwtFromEmailModel(e)
		}

		token, err := h.sessionManager.GenerateJWT(passlink.UserId, emailJwt)
		if err != nil {
			return fmt.Errorf("failed to generate jwt: %w", err)
		}

		cookie, err := h.sessionManager.GenerateCookie(token)
		if err != nil {
			return fmt.Errorf("failed to create session token: %w", err)
		}

		c.Response().Header().Set("X-Session-Lifetime", fmt.Sprintf("%d", cookie.MaxAge))

		if h.cfg.Session.EnableAuthTokenHeader {
			c.Response().Header().Set("X-Auth-Token", token)
		} else {
			c.SetCookie(cookie)
		}

		err = h.auditLogger.CreateWithConnection(tx, c, models.AuditLogPasslinkLoginFinalSucceeded, userModel, nil)
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

		// notify about email verification result. Last step to prevent a trigger and rollback scenario
		if h.cfg.Emails.RequireVerification && wasUnverified {
			var evt events.Event

			if hasEmails {
				evt = events.UserEmailCreate
			} else {
				evt = events.UserCreate
			}

			utils.NotifyUserChange(c, tx, h.persister, evt, userModel.ID)
		}

		return c.JSON(http.StatusOK, dto.PasslinkReturn{
			ID:        passlink.ID.String(),
			CreatedAt: passlink.CreatedAt,
			UserID:    passlink.UserId.String(),
		})
	})

	if businessError != nil {
		return businessError
	}

	return transactionError
}

func (h *PasslinkHandler) GetSessionToken(c echo.Context) jwt.Token {
	var token jwt.Token
	sessionCookie, _ := c.Cookie("hanko")
	// we don't need to check the error, because when the cookie can not be found, the user is not logged in
	if sessionCookie != nil {
		token, _ = h.sessionManager.Verify(sessionCookie.Value)
		// we don't need to check the error, because when the token is not returned, the user is not logged in
	}

	if token == nil {
		authorizationHeader := c.Request().Header.Get("Authorization")
		sessionToken := strings.TrimPrefix(authorizationHeader, "Bearer")
		if strings.TrimSpace(sessionToken) != "" {
			token, _ = h.sessionManager.Verify(sessionToken)
		}
	}

	return token
}

func (h *PasslinkHandler) createRedirectURL(c echo.Context, id uuid.UUID, token string, path string) (string, error) {
	redirect, err := url.Parse(h.URL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL for passlink finalization: %w", err)
	}

	redirect.Path = path

	queryValues := redirect.Query()
	queryValues.Add("plid", id.String())
	queryValues.Add("pltk", token)
	redirect.RawQuery = queryValues.Encode()

	return redirect.String(), nil
}
