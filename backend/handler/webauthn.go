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
}

// NewWebauthnHandler creates a new handler which handles all webauthn related routes
func NewWebauthnHandler(cfg config.WebauthnSettings, persister persistence.Persister, sessionManager session.Manager) (*WebauthnHandler, error) {
	f := false
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName:         cfg.RelyingParty.DisplayName,
		RPID:                  cfg.RelyingParty.Id,
		RPOrigin:              cfg.RelyingParty.Origin,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &f,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Timeout: cfg.Timeout,
		Debug:   false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	return &WebauthnHandler{
		persister:      persister,
		webauthn:       wa,
		sessionManager: sessionManager,
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
	webauthnUser, err := h.getWebauthnUser(h.persister.GetConnection(), uId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if webauthnUser == nil {
		return dto.NewHTTPError(http.StatusBadRequest, "user not found").SetInternal(errors.New(fmt.Sprintf("user %s not found ", uId)))
	}

	t := true
	options, sessionData, err := h.webauthn.BeginRegistration(
		webauthnUser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			RequireResidentKey:      &t,
			ResidentKey:             protocol.ResidentKeyRequirementRequired,
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
			return dto.NewHTTPError(http.StatusBadRequest, "Stored challenge and received challenge do not match").SetInternal(errors.New("sessionData not found"))
		}

		if sessionToken.Subject() != sessionData.UserId.String() {
			return dto.NewHTTPError(http.StatusBadRequest, "Stored challenge and received challenge do not match").SetInternal(errors.New("userId in webauthn.sessionData does not match user session"))
		}

		webauthnUser, err := h.getWebauthnUser(tx, sessionData.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if webauthnUser == nil {
			return dto.NewHTTPError(http.StatusBadRequest).SetInternal(errors.New("user not found"))
		}

		credential, err := h.webauthn.CreateCredential(webauthnUser, *intern.WebauthnSessionDataFromModel(sessionData), request)
		if err != nil {
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

		return c.JSON(http.StatusOK, map[string]string{"credential_id": model.ID, "user_id": webauthnUser.UserId.String()})
	})
}

// BeginAuthentication returns credential assertion options for the WebAuthnAPI.
func (h *WebauthnHandler) BeginAuthentication(c echo.Context) error {
	options, sessionData, err := h.webauthn.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		return fmt.Errorf("failed to create webauthn assertion options: %w", err)
	}

	err = h.persister.GetWebauthnSessionDataPersister().Create(*intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationAuthentication))
	if err != nil {
		return fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	}

	return c.JSON(http.StatusOK, options)
}

// FinishAuthentication validates the WebAuthnAPI response and on success it returns a new session JWT.
func (h *WebauthnHandler) FinishAuthentication(c echo.Context) error {
	request, err := protocol.ParseCredentialRequestResponse(c.Request())
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userId, err := uuid.FromBytes(request.Response.UserHandle)
	if err != nil {
		return dto.NewHTTPError(http.StatusBadRequest, "failed to parse userHandle as uuid").SetInternal(err)
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
			return dto.NewHTTPError(http.StatusUnauthorized, "Stored challenge and received challenge do not match").SetInternal(errors.New("sessionData not found"))
		}

		webauthnUser, err := h.getWebauthnUser(tx, userId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if webauthnUser == nil {
			return dto.NewHTTPError(http.StatusUnauthorized).SetInternal(errors.New("user not found"))
		}

		model := intern.WebauthnSessionDataFromModel(sessionData)
		credential, err := h.webauthn.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (user webauthn.User, err error) {
			return webauthnUser, nil
		}, *model, request)
		if err != nil {
			return dto.NewHTTPError(http.StatusUnauthorized, "failed to validate assertion").SetInternal(err)
		}

		err = sessionDataPersister.Delete(*sessionData)
		if err != nil {
			return fmt.Errorf("failed to delete assertion session data: %w", err)
		}

		cookie, err := h.sessionManager.GenerateCookie(webauthnUser.UserId)
		if err != nil {
			return fmt.Errorf("failed to create session cookie: %w", err)
		}

		c.SetCookie(cookie)
		return c.JSON(http.StatusOK, map[string]string{"credential_id": base64.RawURLEncoding.EncodeToString(credential.ID), "user_id": webauthnUser.UserId.String()})
	})
}

func (h WebauthnHandler) getWebauthnUser(connection *pop.Connection, userId uuid.UUID) (*intern.WebauthnUser, error) {
	user, err := h.persister.GetUserPersisterWithConnection(connection).Get(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, nil
	}

	credentials, err := h.persister.GetWebauthnCredentialPersisterWithConnection(connection).GetFromUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webauthn credentials: %w", err)
	}

	return intern.NewWebauthnUser(*user, credentials), nil
}
