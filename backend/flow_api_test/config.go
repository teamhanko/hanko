package flow_api_test

const (
	FlowOptionPasscodeOnly      uint8 = 1 << iota // passcodes enabled
	FlowOptionSecondFactorFlow                    // use second factor flow (email verification and passwords must be enabled)
	FlowOptionEmailVerification                   // enable email verification
	FlowOptionPasswords                           // enable passwords
)

type FlowConfig struct {
	FlowOption uint8 `json:"flow_option"`
}

func (c *FlowConfig) isEnabled(option uint8) bool {
	return c.FlowOption&option != 0
}
func (c *FlowConfig) IsValid() bool {
	validConfigurations := []uint8{
		FlowOptionPasscodeOnly,
		FlowOptionEmailVerification,
		FlowOptionPasswords,
		FlowOptionEmailVerification | FlowOptionPasswords,
		FlowOptionSecondFactorFlow | FlowOptionEmailVerification | FlowOptionPasswords,
	}

	for _, validOption := range validConfigurations {
		if c.FlowOption == validOption {
			return true
		}
	}

	return false
}

var myFlowConfig FlowConfig
