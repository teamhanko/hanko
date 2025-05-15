package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/persistence"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gobuffalo/nulls"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PatchMetadata struct {
	shared.Action
}

func (a PatchMetadata) GetName() flowpilot.ActionName {
	return shared.ActionPatchMetadata
}

func (a PatchMetadata) GetDescription() string {
	return "Patch the (unsafe) metadata of the user"
}

func (a PatchMetadata) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.JSONInput("patch_metadata").Required(true).Hidden(false))
}

func (a PatchMetadata) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	patchMetadata := c.Input().Get("patch_metadata").String()

	looksLikeObject := strings.HasPrefix(patchMetadata, "{") && strings.HasSuffix(patchMetadata, "}")
	if patchMetadata == "" || (patchMetadata != "null" && !looksLikeObject) {
		c.Input().SetError(
			"patch_metadata",
			shared.ErrorInvalidMetadata.Wrap(errors.New("patch metadata must be an object or null")))
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	userMetadataModel, err := deps.Persister.GetUserMetadataPersister().Get(userModel.ID)
	if err != nil {
		return fmt.Errorf("could not fetch user metadata: %w", err)
	}

	if patchMetadata == "null" {
		userMetadataModel.Unsafe = nulls.String{}
	} else {
		currentUnsafeMetadata := "{}"
		if userMetadataModel.Unsafe.Valid {
			currentUnsafeMetadata = userMetadataModel.Unsafe.String
		}

		patchedUnsafeMetadataBytes, err := jsonpatch.MergePatch(
			[]byte(currentUnsafeMetadata),
			[]byte(patchMetadata),
		)
		if err != nil {
			return fmt.Errorf("could patch unsafe metadata: %w", err)
		}

		if !json.Valid(patchedUnsafeMetadataBytes) {
			return fmt.Errorf("invalid metadata JSON after applying merge patch")
		}

		userMetadataModel.Unsafe = nulls.String{
			Valid:  true,
			String: string(patchedUnsafeMetadataBytes),
		}
	}

	err = deps.Persister.GetUserMetadataPersister().Update(userMetadataModel)
	if err != nil {
		if persistence.IsMetadataLimitExceededError(err) {
			c.Input().SetError(
				"patch_metadata",
				shared.ErrorInvalidMetadata.Wrap(errors.New("metadata must not exceed character limit (3000)")))
			return c.Error(flowpilot.ErrorFormDataInvalid)
		}

		return fmt.Errorf("could not save metadata: %w", err)
	}

	return c.Continue(shared.StateProfileInit)
}
