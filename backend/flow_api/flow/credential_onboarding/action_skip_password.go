package credential_onboarding

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type SkipPassword struct {
	shared.Action
}

func (a SkipPassword) GetName() flowpilot.ActionName {
	return shared.ActionSkip
}

func (a SkipPassword) GetDescription() string {
	return "Skip"
}

func (a SkipPassword) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)
	switch c.GetFlowName() {
	case "registration":
		if !deps.Cfg.Password.Optional || !deps.Cfg.Email.RequireVerification {
			c.SuspendAction()
		}

		if c.GetFlowPath().HasFragment("registration_method_chooser") {
			c.SuspendAction()
		}
	}
}
func (a SkipPassword) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userId := uuid.FromStringOrNil(c.Stash().Get("user_id").String())
	user, err := deps.Persister.GetUserPersister().Get(userId)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if len(user.WebauthnCredentials) > 0 {
		return c.EndSubFlow()
	}

	return c.ContinueFlow(shared.StateOnboardingCreatePasskey)

}

func (a SkipPassword) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
