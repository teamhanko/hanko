package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

type WebauthnHandler struct {
	persister      persistence.Persister
	webauthn       *webauthn.WebAuthn
	sessionManager session.Manager
	cfg            *config.Config
	auditLogClient auditlog.Client
}

// NewWebauthnHandler creates a new handler which handles all webauthn related routes
func NewWebauthnHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogClient auditlog.Client) (*WebauthnHandler, error) {
	f := false
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName:         cfg.Webauthn.RelyingParty.DisplayName,
		RPID:                  cfg.Webauthn.RelyingParty.Id,
		RPOrigin:              cfg.Webauthn.RelyingParty.Origin,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &f,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Timeout: cfg.Webauthn.Timeout,
		Debug:   false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	return &WebauthnHandler{
		persister:      persister,
		webauthn:       wa,
		sessionManager: sessionManager,
		cfg:            cfg,
		auditLogClient: auditLogClient,
	}, nil
}

// BeginRegistration returns credential creation options for the WebAuthnAPI. It expects a valid session JWT in the request.
func (h *WebauthnHandler) BeginRegistration(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}
	uId, err := uuid.FromString(sessionToken.Subject())
	if err != nil {
		return fmt.Errorf("failed to parse userId from JWT subject:%w", err)
	}
	webauthnUser, user, err := h.getWebauthnUser(h.persister.GetConnection(), uId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if webauthnUser == nil {
		err = h.auditLogClient.Create(c, models.AuditLogWebAuthnRegistrationInitFailed, nil, fmt.Errorf("unknown user"))
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}
		return dto.NewHTTPError(http.StatusBadRequest, "user not found").SetInternal(errors.New(fmt.Sprintf("user %s not found ", uId)))
	}

	t := true
	options, sessionData, err := h.webauthn.BeginRegistration(
		webauthnUser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			RequireResidentKey:      &t,
			ResidentKey:             protocol.ResidentKeyRequirementPreferred,
			UserVerification:        protocol.VerificationRequired,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
		// don't set the excludeCredentials list, so an already registered device can be re-registered
	)

	if err != nil {
		return fmt.Errorf("failed to create webauthn creation options: %w", err)
	}

	err = h.persister.GetWebauthnSessionDataPersister().Create(*intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationRegistration))
	if err != nil {
		return fmt.Errorf("failed to store creation options session data: %w", err)
	}

	err = h.auditLogClient.Create(c, models.AuditLogWebAuthnRegistrationInitSucceeded, user, nil)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return c.JSON(http.StatusOK, options)
}

// FinishRegistration validates the WebAuthnAPI response and associates the credential with the user. It expects a valid session JWT in the request.
// The session JWT must be associated to the same user who requested the credential creation options.
func (h *WebauthnHandler) FinishRegistration(c echo.Context) error {
	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}
	request, err := protocol.ParseCredentialCreationResponse(c.Request())
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return h.persister.Transaction(func(tx *pop.Connection) error {
		sessionDataPersister := h.persister.GetWebauthnSessionDataPersisterWithConnection(tx)
		sessionData, err := sessionDataPersister.GetByChallenge(request.Response.CollectedClientData.Challenge)
		if err != nil {
			return fmt.Errorf("failed to get webauthn registration session data: %w", err)
		}

		if sessionData != nil && sessionData.Operation != models.WebauthnOperationRegistration {
			sessionData = nil
		}

		if sessionData == nil {
			err = h.auditLogClient.Create(c, models.AuditLogWebAuthnRegistrationFailed, nil, fmt.Errorf("received unkown challenge"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			return dto.NewHTTPError(http.StatusBadRequest, "Stored challenge and received challenge do not match").SetInternal(errors.New("sessionData not found"))
		}

		if sessionToken.Subject() != sessionData.UserId.String() {
			err = h.auditLogClient.Create(c, models.AuditLogWebAuthnRegistrationFailed, nil, fmt.Errorf("user session does not match sessionData subject"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			return dto.NewHTTPError(http.StatusBadRequest, "Stored challenge and received challenge do not match").SetInternal(errors.New("userId in webauthn.sessionData does not match user session"))
		}

		webauthnUser, user, err := h.getWebauthnUser(tx, sessionData.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if webauthnUser == nil {
			err = h.auditLogClient.Create(c, models.AuditLogWebAuthnRegistrationFailed, nil, fmt.Errorf("unkown user"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			return dto.NewHTTPError(http.StatusBadRequest).SetInternal(errors.New("user not found"))
		}

		credential, err := h.webauthn.CreateCredential(webauthnUser, *intern.WebauthnSessionDataFromModel(sessionData), request)
		if err != nil {
			err = h.auditLogClient.Create(c, models.AuditLogWebAuthnRegistrationFailed, user, fmt.Errorf("attestation validation failed"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			return dto.NewHTTPError(http.StatusBadRequest, "Failed to validate attestation").SetInternal(err)
		}

		model := intern.WebauthnCredentialToModel(credential, sessionData.UserId)
		err = h.persister.GetWebauthnCredentialPersisterWithConnection(tx).Create(*model)
		if err != nil {
			return fmt.Errorf("failed to store webauthn credential: %w", err)
		}

		err = sessionDataPersister.Delete(*sessionData)
		if err != nil {
			c.Logger().Errorf("failed to delete attestation session data: %w", err)
		}

		err = h.auditLogClient.Create(c, models.AuditLogWebAuthnRegistrationSucceeded, user, nil)
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

		return c.JSON(http.StatusOK, map[string]string{"credential_id": model.ID, "user_id": webauthnUser.UserId.String()})
	})
}

type BeginAuthenticationBody struct {
	UserID *string `json:"user_id" validate:"uuid4"`
}

// BeginAuthentication returns credential assertion options for the WebAuthnAPI.
func (h *WebauthnHandler) BeginAuthentication(c echo.Context) error {
	var request BeginAuthenticationBody

	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return dto.ToHttpError(err)
	}

	var options *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	var user *models.User
	if request.UserID != nil {
		// non discoverable login initialization
		userId, err := uuid.FromString(*request.UserID)
		if err != nil {
			err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationInitFailed, nil, fmt.Errorf("user_id is not a uuid"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			return dto.NewHTTPError(http.StatusBadRequest, "failed to parse UserID as uuid").SetInternal(err)
		}
		var webauthnUser *intern.WebauthnUser
		webauthnUser, user, err = h.getWebauthnUser(h.persister.GetConnection(), userId) // TODO:
		if err != nil {
			return dto.NewHTTPError(http.StatusInternalServerError).SetInternal(fmt.Errorf("failed to get user: %w", err))
		}
		if webauthnUser == nil {
			err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationInitFailed, nil, fmt.Errorf("unkown user"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			return dto.NewHTTPError(http.StatusBadRequest, "user not found")
		}

		if len(webauthnUser.WebAuthnCredentials()) > 0 {
			options, sessionData, err = h.webauthn.BeginLogin(webauthnUser, webauthn.WithUserVerification(protocol.VerificationRequired))
			if err != nil {
				return fmt.Errorf("failed to create webauthn assertion options: %w", err)
			}
		}
	}
	if options == nil && sessionData == nil {
		var err error
		options, sessionData, err = h.webauthn.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
		if err != nil {
			return fmt.Errorf("failed to create webauthn assertion options for discoverable login: %w", err)
		}
	}

	err := h.persister.GetWebauthnSessionDataPersister().Create(*intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationAuthentication))
	if err != nil {
		return fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	}

	// Remove all transports, because of a bug in android and windows where the internal authenticator gets triggered,
	// when the transports array contains the type 'internal' although the credential is not available on the device.
	for i, _ := range options.Response.AllowedCredentials {
		options.Response.AllowedCredentials[i].Transport = nil
	}

	err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationInitSucceeded, user, nil)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return c.JSON(http.StatusOK, options)
}

// FinishAuthentication validates the WebAuthnAPI response and on success it returns a new session JWT.
func (h *WebauthnHandler) FinishAuthentication(c echo.Context) error {
	request, err := protocol.ParseCredentialRequestResponse(c.Request())
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		sessionDataPersister := h.persister.GetWebauthnSessionDataPersisterWithConnection(tx)
		sessionData, err := sessionDataPersister.GetByChallenge(request.Response.CollectedClientData.Challenge)
		if err != nil {
			return fmt.Errorf("failed to get webauthn assertion session data: %w", err)
		}

		if sessionData != nil && sessionData.Operation != models.WebauthnOperationAuthentication {
			sessionData = nil
		}

		if sessionData == nil {
			err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationFailed, nil, fmt.Errorf("received unkown challenge"))
			if err != nil {
				return fmt.Errorf("failed to create audit log: %w", err)
			}
			return dto.NewHTTPError(http.StatusUnauthorized, "Stored challenge and received challenge do not match").SetInternal(errors.New("sessionData not found"))
		}

		model := intern.WebauthnSessionDataFromModel(sessionData)

		var credential *webauthn.Credential
		var webauthnUser *intern.WebauthnUser
		var user *models.User
		if sessionData.UserId.IsNil() {
			// Discoverable Login
			userId, err := uuid.FromBytes(request.Response.UserHandle)
			if err != nil {
				return dto.NewHTTPError(http.StatusBadRequest, "failed to parse userHandle as uuid").SetInternal(err)
			}
			webauthnUser, user, err = h.getWebauthnUser(tx, userId)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			if webauthnUser == nil {
				err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationFailed, nil, fmt.Errorf("unkown user"))
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
				return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("user not found"))
			}

			credential, err = h.webauthn.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (user webauthn.User, err error) {
				return webauthnUser, nil
			}, *model, request)
			if err != nil {
				err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationFailed, user, fmt.Errorf("assertion validation failed"))
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
				return dto.NewHTTPError(http.StatusUnauthorized, "failed to validate assertion").SetInternal(err)
			}
		} else {
			// non discoverable Login
			webauthnUser, user, err = h.getWebauthnUser(tx, sessionData.UserId)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}
			if webauthnUser == nil {
				err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationFailed, nil, fmt.Errorf("unkown user"))
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
				return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("user not found"))
			}
			credential, err = h.webauthn.ValidateLogin(webauthnUser, *model, request)
			if err != nil {
				err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationFailed, user, fmt.Errorf("assertion validation failed"))
				if err != nil {
					return fmt.Errorf("failed to create audit log: %w", err)
				}
				return dto.NewHTTPError(http.StatusUnauthorized, "failed to validate assertion").SetInternal(err)
			}
		}

		err = sessionDataPersister.Delete(*sessionData)
		if err != nil {
			return fmt.Errorf("failed to delete assertion session data: %w", err)
		}

		token, err := h.sessionManager.GenerateJWT(webauthnUser.UserId)
		if err != nil {
			return fmt.Errorf("failed to generate jwt: %w", err)
		}

		cookie, err := h.sessionManager.GenerateCookie(token)
		if err != nil {
			return fmt.Errorf("failed to create session cookie: %w", err)
		}

		c.SetCookie(cookie)

		if h.cfg.Session.EnableAuthTokenHeader {
			c.Response().Header().Set("X-Auth-Token", token)
		}

		err = h.auditLogClient.Create(c, models.AuditLogWebAuthnAuthenticationSucceeded, user, nil)
		if err != nil {
			return fmt.Errorf("failed to create audit log: %w", err)
		}

		return c.JSON(http.StatusOK, map[string]string{"credential_id": base64.RawURLEncoding.EncodeToString(credential.ID), "user_id": webauthnUser.UserId.String()})
	})
}

func (h WebauthnHandler) getWebauthnUser(connection *pop.Connection, userId uuid.UUID) (*intern.WebauthnUser, *models.User, error) {
	user, err := h.persister.GetUserPersisterWithConnection(connection).Get(userId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, nil, nil
	}

	credentials, err := h.persister.GetWebauthnCredentialPersisterWithConnection(connection).GetFromUser(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get webauthn credentials: %w", err)
	}

	return intern.NewWebauthnUser(*user, credentials), user, nil
}
