package shared

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/dto/intern"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type VerifyAttestationResponse struct {
	Action
}

func (h VerifyAttestationResponse) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	if !c.Stash().Get(StashPathWebauthnSessionDataID).Exists() {
		return errors.New("webauthn_session_data_id does not exist in the stash")
	}

	sessionDataID, err := uuid.FromString(c.Stash().Get(StashPathWebauthnSessionDataID).String())
	if err != nil {
		return fmt.Errorf("failed to parse webauthn_session_data_id: %w", err)
	}

	userID, err := uuid.FromString(c.Stash().Get(StashPathUserID).String())
	if err != nil {
		return fmt.Errorf("failed to parse user_id into a uuid: %w", err)
	}

	username := c.Stash().Get(StashPathUsername).String()
	email := c.Stash().Get(StashPathEmail).String()

	params := services.VerifyAttestationResponseParams{
		Tx:            deps.Tx,
		SessionDataID: sessionDataID,
		PublicKey:     c.Input().Get("public_key").String(),
		UserID:        userID,
		Email:         &email,
		Username:      &username,
	}

	credential, err := deps.WebauthnService.VerifyAttestationResponse(params)
	if err != nil {
		if errors.Is(err, services.ErrInvalidWebauthnCredential) {
			c.SetFlowError(ErrorPasskeyInvalid.Wrap(err))
			return nil
		}

		return fmt.Errorf("failed to verify attestation response: %w", err)
	}

	mfaOnly := c.Stash().Get(StashPathCreateMFAOnlyCredential).Bool()
	credentialModel := intern.WebauthnCredentialToModel(credential, userID, false, false, mfaOnly, deps.AuthenticatorMetadata)
	err = c.Stash().Set(fmt.Sprintf("%s.-1", StashPathWebauthnCredentials), credentialModel)
	if err != nil {
		return fmt.Errorf("failed to set webauthn_credential to the stash: %w", err)
	}

	err = c.Stash().Set(StashPathUserHasWebauthnCredential, true)
	if err != nil {
		return fmt.Errorf("failed to set user_has_webauthn_credential to the stash: %w", err)
	}

	if mfaOnly {
		err = c.Stash().Set(StashPathUserHasSecurityKey, true)
		if err != nil {
			return fmt.Errorf("failed to set user_has_security_key to the stash: %w", err)
		}
	} else {
		err = c.Stash().Set(StashPathUserHasPasskey, true)
		if err != nil {
			return fmt.Errorf("failed to set user_has_passkey to the stash: %w", err)
		}
	}

	return nil
}
