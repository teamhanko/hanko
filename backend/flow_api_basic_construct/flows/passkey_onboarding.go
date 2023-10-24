package flows

import (
	"github.com/go-webauthn/webauthn/protocol"
	webauthnLib "github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/actions"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/session"
	"time"
)

func NewPasskeyOnboardingSubFlow(cfg config.Config, persister persistence.Persister, userService services.User, sessionManager session.Manager, httpContext echo.Context) (flowpilot.SubFlow, error) {
	// TODO:
	f := false
	wa, err := webauthnLib.New(&webauthnLib.Config{
		RPID:                  cfg.Webauthn.RelyingParty.Id,
		RPDisplayName:         cfg.Webauthn.RelyingParty.DisplayName,
		RPOrigins:             cfg.Webauthn.RelyingParty.Origins,
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &f,
			ResidentKey:        protocol.ResidentKeyRequirementDiscouraged,
			UserVerification:   protocol.VerificationRequired,
		},
		Debug: false,
		Timeouts: webauthnLib.TimeoutsConfig{
			Login: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(cfg.Webauthn.Timeout) * time.Millisecond,
			},
			Registration: webauthnLib.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(cfg.Webauthn.Timeout) * time.Millisecond,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return flowpilot.NewSubFlow().
		State(common.StateOnboardingCreatePasskey, actions.NewGetWACreationOptions(cfg, persister, wa), actions.NewSkip(cfg)).
		State(common.StateOnboardingVerifyPasskeyAttestation, actions.NewSendWAAttestationResponse(cfg, persister, wa, userService, sessionManager, httpContext)).
		Build()
}
