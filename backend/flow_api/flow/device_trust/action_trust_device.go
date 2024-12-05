package device_trust

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type TrustDevice struct {
	shared.Action
}

func (a TrustDevice) GetName() flowpilot.ActionName {
	return shared.ActionTrustDevice
}

func (a TrustDevice) GetDescription() string {
	return "Trust this device, to skip MFA on subsequent logins."
}

func (a TrustDevice) Initialize(c flowpilot.InitializationContext) {}

func (a TrustDevice) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set(shared.StashPathDeviceTrustGranted, true); err != nil {
		return fmt.Errorf("failed to set device_trust_granted to the stash: %w", err)
	}

	c.PreventRevert()

	return c.Continue()
}
