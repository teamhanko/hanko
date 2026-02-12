package credential_usage

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/teamhanko/hanko/backend/v2/dto/webhook"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/rate_limiter"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
	"github.com/teamhanko/hanko/backend/v2/webhooks/utils"
)

type ReSendPasscode struct {
	shared.Action
}

func (a ReSendPasscode) GetName() flowpilot.ActionName {
	return shared.ActionResendPasscode
}

func (a ReSendPasscode) GetDescription() string {
	return "Send the passcode email again."
}

func (a ReSendPasscode) Initialize(_ flowpilot.InitializationContext) {}

func (a ReSendPasscode) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if !c.Stash().Get(shared.StashPathEmail).Exists() {
		return errors.New("email has not been stashed")
	}

	if !c.Stash().Get(shared.StashPathPasscodeTemplate).Exists() {
		return errors.New("passcode_template has not been stashed")
	}

	if deps.Cfg.RateLimiter.Enabled {
		rateLimitKey := rate_limiter.CreateRateLimitPasscodeKey(deps.HttpContext.RealIP(), c.Stash().Get(shared.StashPathEmail).String())
		resendAfterSeconds, ok, err := rate_limiter.Limit2(deps.PasscodeRateLimiter, rateLimitKey)
		if err != nil {
			return fmt.Errorf("rate limiter failed: %w", err)
		}

		if !ok {
			err = c.Payload().Set("resend_after", resendAfterSeconds)
			if err != nil {
				return fmt.Errorf("failed to set a value for resend_after to the payload: %w", err)
			}
			return c.Error(shared.ErrorRateLimitExceeded.Wrap(fmt.Errorf("rate limit exceeded for: %s", rateLimitKey)))
		}
	}

	passcodeTemplate := c.Stash().Get(shared.StashPathPasscodeTemplate).String()

	sendParams := services.SendPasscodeParams{
		Template:     passcodeTemplate,
		EmailAddress: c.Stash().Get(shared.StashPathEmail).String(),
		Language:     deps.HttpContext.Request().Header.Get("X-Language"),
	}
	passcodeResult, err := deps.PasscodeService.SendPasscode(deps.Tx, sendParams)
	if err != nil {
		return fmt.Errorf("passcode service failed: %w", err)
	}

	webhookData := webhook.EmailSend{
		Subject:          passcodeResult.Subject,
		BodyPlain:        passcodeResult.BodyPlain,
		Body:             passcodeResult.BodyHTML,
		ToEmailAddress:   sendParams.EmailAddress,
		DeliveredByHanko: deps.Cfg.EmailDelivery.Enabled,
		AcceptLanguage:   sendParams.Language,
		Language:         sendParams.Language,
		Type:             passcodeTemplate,
	}

	if slices.Contains(
		[]string{
			shared.PasscodeTemplateEmailRegistrationAttempted,
			shared.PasscodeTemplateEmailLoginAttempted,
		}, passcodeTemplate) {
		webhookData.Data = webhook.PasscodeData{
			ServiceName: deps.Cfg.Service.Name,
		}
	} else {
		webhookData.Data = webhook.PasscodeData{
			ServiceName: deps.Cfg.Service.Name,
			OtpCode:     passcodeResult.Code,
			TTL:         deps.Cfg.Email.PasscodeTtl,
			ValidUntil:  passcodeResult.PasscodeModel.CreatedAt.Add(time.Duration(deps.Cfg.Email.PasscodeTtl) * time.Second).UTC().Unix(),
		}
	}

	err = utils.TriggerWebhooks(deps.HttpContext, deps.Tx, events.EmailSend, webhookData)
	if err != nil {
		return fmt.Errorf("failed to trigger webhook: %w", err)
	}

	err = c.Stash().Set(shared.StashPathPasscodeID, passcodeResult.PasscodeModel.ID)
	if err != nil {
		return fmt.Errorf("failed to set passcode_id to stash: %w", err)
	}

	return c.Continue(c.GetCurrentState())
}
