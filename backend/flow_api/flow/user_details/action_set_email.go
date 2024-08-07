package user_details

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type EmailAddressSet struct {
	shared.Action
}

func (a EmailAddressSet) GetName() flowpilot.ActionName {
	return shared.ActionEmailAddressSet
}

func (a EmailAddressSet) GetDescription() string {
	return "Set a new email address."
}

func (a EmailAddressSet) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	c.AddInputs(flowpilot.StringInput("email").
		Required(!deps.Cfg.Email.Optional).
		MaxLength(deps.Cfg.Email.MaxLength).
		Preserve(true).
		TrimSpace(true).
		LowerCase(true))
}

func (a EmailAddressSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	email := c.Input().Get("email").String()

	err := c.Stash().Set(shared.StashPathEmail, email)
	if err != nil {
		return fmt.Errorf("failed to stash email address: %w", err)
	}

	existingEmail, err := deps.Persister.GetEmailPersister().FindByAddress(email)
	if err != nil {
		return fmt.Errorf("failed to get email from db: %w", err)
	}

	if deps.Cfg.Email.RequireVerification {
		// Email verification is enabled. Send an email regardless of whether the email address exists, but select the
		// appropriate passcode template beforehand.
		if existingEmail != nil {
			err = c.Stash().Set(shared.StashPathPasscodeTemplate, "email_registration_attempted") // "email_verification"
			if err != nil {
				return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
			}
		} else {
			err = c.Stash().Set(shared.StashPathPasscodeTemplate, "email_verification")
			if err != nil {
				return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
			}
		}

		if err = c.Stash().Set(shared.StashPathLoginOnboardingCreateEmail, true); err != nil {
			return fmt.Errorf("failed to set login_onboarding_create_email to the stash: %w", err)
		}

		return c.Continue(shared.StatePasscodeConfirmation)
	}

	// Email verification is turned off, hence we can display an error if the email already exists, or continue the flow
	// without passcode verification otherwise.
	if existingEmail != nil {
		c.Input().SetError("email", shared.ErrorEmailAlreadyExists)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	if err = c.Stash().Set(shared.StashPathLoginOnboardingCreateEmail, true); err != nil {
		return fmt.Errorf("failed to set login_onboarding_create_email to the stash: %w", err)
	}

	c.PreventRevert()

	return c.Continue()
}
