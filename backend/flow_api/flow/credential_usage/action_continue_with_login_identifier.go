package credential_usage

import (
	"errors"
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"regexp"
	"strings"
)

type ContinueWithLoginIdentifier struct {
	shared.Action
}

func (a ContinueWithLoginIdentifier) GetName() flowpilot.ActionName {
	return shared.ActionContinueWithLoginIdentifier
}

func (a ContinueWithLoginIdentifier) GetDescription() string {
	return "Enter an identifier to login."
}

func (a ContinueWithLoginIdentifier) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	emailEnabled := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseAsLoginIdentifier
	usernameEnabled := deps.Cfg.Username.Enabled && deps.Cfg.Username.UseAsLoginIdentifier

	var input flowpilot.Input
	if usernameEnabled && emailEnabled {
		input = flowpilot.StringInput("identifier").
			MaxLength(255)
	} else if emailEnabled {
		input = flowpilot.EmailInput("email").
			MaxLength(deps.Cfg.Email.MaxLength).
			MinLength(3)
	} else if usernameEnabled {
		input = flowpilot.StringInput("username").
			MaxLength(deps.Cfg.Username.MaxLength).
			MinLength(deps.Cfg.Username.MinLength)
	}

	if input != nil {
		c.AddInputs(input.
			Required(true).
			TrimSpace(true).
			LowerCase(true))
	}

	if !deps.Cfg.Password.Enabled &&
		!deps.Cfg.Email.UseForAuthentication &&
		!(emailEnabled && deps.Cfg.Saml.Enabled && len(deps.SamlService.Providers()) > 0) {
		c.SuspendAction()
	}

	if !emailEnabled && !usernameEnabled {
		c.SuspendAction()
	}
}

func (a ContinueWithLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	identifierInputName, identifierInputValue, treatIdentifierAsEmail := a.analyzeIdentifierInputs(c)

	if err := c.Stash().Set(shared.StashPathUserIdentification, identifierInputValue); err != nil {
		return fmt.Errorf("failed to set user_identification to stash: %w", err)
	}

	if len(identifierInputValue) == 0 {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	var userModel *models.User

	if treatIdentifierAsEmail {
		// User has submitted an email address.

		var err error

		userModel, err = deps.Persister.GetUserPersister().GetByEmailAddress(identifierInputValue)
		if err != nil {
			return err
		}

		if err = c.Stash().Set(shared.StashPathEmail, identifierInputValue); err != nil {
			return fmt.Errorf("failed to set email to stash: %w", err)
		}

		if userModel != nil {
			emailModel := userModel.GetEmailByAddress(identifierInputValue)

			if emailModel != nil && emailModel.UserID != nil {
				err = c.Stash().Set(shared.StashPathUserID, emailModel.UserID.String())
				if err != nil {
					return fmt.Errorf("failed to set user_id to the stash: %w", err)
				}
			}
		}

		if deps.Cfg.Saml.Enabled {
			domain := strings.Split(identifierInputValue, "@")[1]
			if provider, err := deps.SamlService.GetProviderByDomain(domain); err == nil && provider != nil {
				authUrl, err := deps.SamlService.GetAuthUrl(provider, deps.Cfg.Saml.DefaultRedirectUrl, true)

				if err != nil {
					return fmt.Errorf("failed to get auth url: %w", err)
				}

				_ = c.Payload().Set("redirect_url", authUrl)

				return c.Continue(shared.StateThirdParty)
			}
		}
	} else {
		// User has submitted a username.
		var err error

		userModel, err = deps.Persister.GetUserPersister().GetByUsername(identifierInputValue)
		if err != nil {
			return fmt.Errorf("failed to get user by username from db: %w", err)
		}

		if userModel == nil {
			flowInputError := shared.ErrorUnknownUsername
			err = deps.AuditLogger.CreateWithConnection(
				deps.Tx,
				deps.HttpContext,
				models.AuditLogLoginFailure,
				nil,
				flowInputError,
				auditlog.Detail("flow_id", c.GetFlowID()))

			if err != nil {
				return fmt.Errorf("could not create audit log: %w", err)
			}

			c.Input().SetError(identifierInputName, flowInputError)
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}

		if err = c.Stash().Set(shared.StashPathUsername, identifierInputValue); err != nil {
			return fmt.Errorf("failed to set username to stash: %w", err)
		}

		err = c.Stash().Set(shared.StashPathUserID, userModel.ID.String())
		if err != nil {
			return fmt.Errorf("failed to set user_id to the stash: %w", err)
		}
		if primaryEmailModel := userModel.Emails.GetPrimary(); primaryEmailModel != nil {
			if err = c.Stash().Set(shared.StashPathEmail, primaryEmailModel.Address); err != nil {
				return fmt.Errorf("failed to set email to stash: %w", err)
			}
		}
	}

	if userModel != nil {
		_ = c.Stash().Set(shared.StashPathUserHasPassword, userModel.PasswordCredential != nil)
		_ = c.Stash().Set(shared.StashPathUserHasPasskey, len(userModel.GetPasskeys()) > 0)
		_ = c.Stash().Set(shared.StashPathUserHasWebauthnCredential, len(userModel.WebauthnCredentials) > 0)
		_ = c.Stash().Set(shared.StashPathUserHasUsername, userModel.GetUsername() != nil)
		_ = c.Stash().Set(shared.StashPathUserHasEmails, len(userModel.Emails) > 0)
		_ = c.Stash().Set(shared.StashPathUserHasOTPSecret, userModel.OTPSecret != nil)
		_ = c.Stash().Set(shared.StashPathUserHasSecurityKey, len(userModel.GetSecurityKeys()) > 0)
	}

	if !treatIdentifierAsEmail && userModel != nil && !deps.Cfg.Password.Enabled && userModel.Emails.GetPrimary() == nil {
		// The user has entered a username of an existing user, but passwords are disabled, and the user does not have
		// an email address to send the passcode.
		return c.Error(flowpilot.ErrorFlowDiscontinuity.Wrap(errors.New("user has no email address and passwords are disabled")))
	}

	if deps.Cfg.Email.UseForAuthentication && deps.Cfg.Password.Enabled {
		// Both passcode and password authentication are enabled.
		if treatIdentifierAsEmail || (!treatIdentifierAsEmail && userModel != nil && userModel.Emails.GetPrimary() != nil) {
			// The user has entered either an email address, or a username for an existing user who has an email address.
			return c.Continue(shared.StateLoginMethodChooser)
		}

		// Either no email was entered or the username does not correspond to an email, passwords are enabled.
		return c.Continue(shared.StateLoginPassword)
	}

	if deps.Cfg.Email.UseForAuthentication {
		// Only passcode authentication is enabled; the user must use a passcode.

		// Set the login method for audit logging purposes.
		if err := c.Stash().Set(shared.StashPathLoginMethod, "passcode"); err != nil {
			return fmt.Errorf("failed to set login_method to stash: %w", err)
		}

		if c.Stash().Get(shared.StashPathUserID).Exists() {
			if err := c.Stash().Set(shared.StashPathPasscodeTemplate, "login"); err != nil {
				return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
			}
		} else {
			if err := c.Stash().Set(shared.StashPathPasscodeTemplate, "email_login_attempted"); err != nil {
				return fmt.Errorf("failed to set passcode_template to the stash: %w", err)
			}
		}

		return c.Continue(shared.StatePasscodeConfirmation)
	}

	if deps.Cfg.Password.Enabled {
		// Only password authentication is enabled; the user must use a password.
		return c.Continue(shared.StateLoginPassword)
	}

	return c.Error(flowpilot.ErrorFlowDiscontinuity.Wrap(errors.New("no authentication method enabled")))
}

// analyzeIdentifierInputs determines if an input value has been provided for 'identifier', 'email', or 'username',
// according to the configuration. Also adds an input error to the expected input field, if the value is missing.
// Returns the related input field name, the provided value, and a flag, indicating if the value should be treated as
// an email (and not as a username).
func (a ContinueWithLoginIdentifier) analyzeIdentifierInputs(c flowpilot.ExecutionContext) (name, value string, treatAsEmail bool) {
	deps := a.GetDeps(c)
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	emailEnabled := deps.Cfg.Email.Enabled && deps.Cfg.Email.UseAsLoginIdentifier
	usernameEnabled := deps.Cfg.Username.Enabled && deps.Cfg.Username.UseAsLoginIdentifier

	if emailEnabled && usernameEnabled {
		// analyze the 'identifier' input field
		name = "identifier"
		value = c.Input().Get(name).String()
		treatAsEmail = emailPattern.MatchString(value)
	} else if emailEnabled {
		// analyze the 'email' input field
		name = "email"
		value = c.Input().Get(name).String()
		treatAsEmail = true
	} else if usernameEnabled {
		// analyze the 'username' input field
		name = "username"
		value = c.Input().Get(name).String()
		treatAsEmail = false
	}

	// If no value could not be determined, set an error for the missing input
	if len(value) == 0 && len(name) > 0 {
		c.Input().SetError(name, flowpilot.ErrorValueMissing)
	}

	return name, value, treatAsEmail
}
