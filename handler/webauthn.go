package handler

import (
	"encoding/base64"
	"fmt"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/dto"
	"github.com/teamhanko/hanko/dto/intern"
	"github.com/teamhanko/hanko/persistence"
	"github.com/teamhanko/hanko/persistence/models"
	"net/http"
)

type WebauthnHandler struct {
	persister *persistence.Persister
	webauthn  *webauthn.WebAuthn
}

func NewWebauthnHandler(cfg config.WebauthnSettings, persister *persistence.Persister) (*WebauthnHandler, error) {
	f := false
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName:         cfg.RelyingParty.DisplayName,
		RPID:                  cfg.RelyingParty.Id,
		RPOrigin:              cfg.RelyingParty.Origins[0], // TODO:
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &f,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Timeout: cfg.Timeouts.Registration, // TODO:
		Debug:   false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	return &WebauthnHandler{
		persister: persister,
		webauthn:  wa,
	}, nil
}

func (h *WebauthnHandler) BeginRegistration(c echo.Context) error {
	cookie, err := c.Cookie("hanko-session")
	if err != nil {
		return fmt.Errorf("failed to get cookie: %w", err)
	}
	uId, err := uuid.FromString(cookie.Value)
	webauthnUser, err := h.getWebauthnUser(uId) // TODO: get user from JWT/Session
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if webauthnUser == nil {
		return c.JSON(http.StatusNotFound, dto.NewApiError(http.StatusNotFound))
	}

	f := false
	options, sessionData, err := h.webauthn.BeginRegistration(
		webauthnUser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			RequireResidentKey:      &f,
			ResidentKey:             protocol.ResidentKeyRequirementPreferred,
			UserVerification:        protocol.VerificationRequired,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
		// don't set the excludeCredentials list, so an already registered device can be re-registered
	)

	if err != nil {
		return fmt.Errorf("failed to create webauthn creation options: %w", err)
	}

	err = h.persister.WebAuthnSessionData.Create(*intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationRegistration))
	if err != nil {
		return fmt.Errorf("failed to store creation options session data: %w", err)
	}

	return c.JSON(http.StatusOK, options)
}

func (h *WebauthnHandler) FinishRegistration(c echo.Context) error {
	cookie, err := c.Cookie("hanko-session")
	if err != nil {
		return fmt.Errorf("failed to get cookie: %w", err)
	}
	request, err := protocol.ParseCredentialCreationResponse(c.Request())
	if err != nil {
		c.Logger().Errorf("%w", err)
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}
	return h.persister.DB.Transaction(func(tx *pop.Connection) error {
		sessionData, err := h.persister.WebAuthnSessionData.WithConnection(tx).GetByChallenge(request.Response.CollectedClientData.Challenge)
		if err != nil {
			return fmt.Errorf("failed to get webauthn registration session data: %w", err)
		}

		if sessionData != nil && sessionData.Operation != models.WebauthnOperationRegistration {
			sessionData = nil
		}

		if sessionData == nil {
			return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
		}

		// TODO: check that userID in sessionData equals user from JWT/Session
		if cookie.Value != sessionData.UserId.String() {
			return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
		}

		user, err := h.persister.User.WithConnection(tx).Get(sessionData.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		credentials, err := h.persister.WebAuthnCredential.WithConnection(tx).GetFromUser(user.ID)
		if err != nil {
			return fmt.Errorf("failed to get webauthn credentials: %w", err)
		}

		webauthnUser := intern.NewWebauthnUser(*user, credentials)

		credential, err := h.webauthn.CreateCredential(webauthnUser, *intern.WebauthnSessionDataFromModel(sessionData), request)
		if err != nil {
			// TODO: log error, should we return the error message given from the lib?
			return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
		}

		model := intern.WebauthnCredentialToModel(credential, sessionData.UserId)
		err = h.persister.WebAuthnCredential.Create(*model)
		if err != nil {
			return fmt.Errorf("failed to store webauthn credential")
		}

		err = h.persister.WebAuthnSessionData.WithConnection(tx).Delete(*sessionData)
		if err != nil {
			c.Logger().Errorf("failed to delete attestation session data: %w", err)
		}

		return c.JSON(http.StatusOK, map[string]string{"credential_id": model.ID})
	})
}

func (h *WebauthnHandler) BeginAuthentication(c echo.Context) error {
	body := WebauthnLoginRequest{}
	err := (&echo.DefaultBinder{}).BindBody(c, &body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	userId, err := uuid.FromString(body.UserId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	webauthnUser, err := h.getWebauthnUser(userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if webauthnUser == nil {
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	var allowedCredentials []protocol.CredentialDescriptor
	for _, credential := range webauthnUser.WebauthnCredentials {
		id, _ := base64.RawURLEncoding.DecodeString(credential.ID)
		credentialDescriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: id,
		}

		allowedCredentials = append(allowedCredentials, credentialDescriptor)
	}

	options, sessionData, err := h.webauthn.BeginLogin(
		webauthnUser,
		webauthn.WithUserVerification(protocol.VerificationRequired),
		webauthn.WithAllowedCredentials(allowedCredentials),
	)

	if err != nil {
		return fmt.Errorf("failed to create webauthn assertion options: %w", err)
	}

	err = h.persister.WebAuthnSessionData.Create(*intern.WebauthnSessionDataToModel(sessionData, models.WebauthnOperationAuthentication))
	if err != nil {
		return fmt.Errorf("failed to store webauthn assertion session data: %w", err)
	}

	return c.JSON(http.StatusOK, options)
}

func (h *WebauthnHandler) FinishAuthentication(c echo.Context) error {
	request, err := protocol.ParseCredentialRequestResponse(c.Request())
	if err != nil {
		c.Logger().Errorf("%w", err)
		return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
	}

	return h.persister.DB.Transaction(func(tx *pop.Connection) error {
		sessionData, err := h.persister.WebAuthnSessionData.WithConnection(tx).GetByChallenge(request.Response.CollectedClientData.Challenge)
		if err != nil {
			return fmt.Errorf("failed to get webauthn assertion session data: %w", err)
		}

		if sessionData != nil && sessionData.Operation != models.WebauthnOperationAuthentication {
			sessionData = nil
		}

		if sessionData == nil {
			return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
		}

		webauthnUser, err := h.getWebauthnUser(sessionData.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if webauthnUser == nil {
			return c.JSON(http.StatusBadRequest, dto.NewApiError(http.StatusBadRequest))
		}

		model := intern.WebauthnSessionDataFromModel(sessionData)
		_, err = h.webauthn.ValidateLogin(webauthnUser, *model, request)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, dto.NewApiError(http.StatusUnauthorized))
		}

		err = h.persister.WebAuthnSessionData.WithConnection(tx).Delete(*sessionData)
		if err != nil {
			return fmt.Errorf("failed to delete assertion session data: %w", err)
		}

		// TODO: create JWT and return it
		cookie := &http.Cookie{
			Name:     "hanko-session",
			Value:    webauthnUser.UserId.String(),
			Domain:   "",
			Secure:   true,
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(cookie)
		return c.String(http.StatusOK, "")
	})
}

func (h WebauthnHandler) getWebauthnUser(userId uuid.UUID) (*intern.WebauthnUser, error) {
	user, err := h.persister.User.Get(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, nil
	}

	credentials, err := h.persister.WebAuthnCredential.GetFromUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webauthn credentials: %w", err)
	}

	return intern.NewWebauthnUser(*user, credentials), nil
}

type WebauthnLoginRequest struct {
	UserId string `json:"user_id"`
}
