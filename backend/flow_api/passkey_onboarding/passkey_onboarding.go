package passkey_onboarding

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/actions"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding/states"
	sharedActions "github.com/teamhanko/hanko/backend/flow_api/shared/actions"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
)

func NewPasskeyOnboardingSubFlow(cfg config.Config, persister persistence.Persister) (flowpilot.SubFlow, error) {
	wa, err := cfg.Webauthn.GetConfig()
	if err != nil {
		return nil, err
	}
	return flowpilot.NewSubFlow().
		State(states.StateOnboardingCreatePasskey, actions.NewGetWACreationOptions(cfg, persister, wa), sharedActions.NewSkip(cfg)).
		State(states.StateOnboardingVerifyPasskeyAttestation, actions.NewSendWAAttestationResponse(persister, wa)).
		Build()
}
