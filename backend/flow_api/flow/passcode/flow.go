package passcode

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

const (
	StatePasscodeConfirmation flowpilot.StateName = "passcode_confirmation"
)

const (
	ActionVerifyPasscode flowpilot.ActionName = "verify_passcode"
	ActionResendPasscode flowpilot.ActionName = "resend_passcode"
)

var SubFlow = flowpilot.NewSubFlow().
	State(StatePasscodeConfirmation, VerifyPasscode{}, ReSendPasscode{}, shared.Back{}).
	BeforeState(StatePasscodeConfirmation, SendPasscode{}).
	MustBuild()

func createRateLimitKey(realIP, email string) string {
	return fmt.Sprintf("%s/%s", realIP, email)
}
