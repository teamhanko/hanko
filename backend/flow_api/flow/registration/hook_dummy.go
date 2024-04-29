package registration

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type Dummy struct {
	shared.Action
}

func (h Dummy) Execute(c flowpilot.HookExecutionContext) error {
	fmt.Println("DUMMY HOOK EXECUTING")
	return nil
}
