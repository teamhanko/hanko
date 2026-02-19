package shared

import "github.com/teamhanko/hanko/backend/v2/flowpilot"

const (
	FlowCapabilities         flowpilot.FlowName = "capabilities"
	FlowCredentialOnboarding flowpilot.FlowName = "credential_onboarding"
	FlowCredentialUsage      flowpilot.FlowName = "credential_usage"
	FlowDeviceTrust          flowpilot.FlowName = "device_trust"
	FlowExchangeToken        flowpilot.FlowName = "exchange_token"
	FlowLogin                flowpilot.FlowName = "login"
	FlowMFACreation          flowpilot.FlowName = "mfa_creation"
	FlowProfile              flowpilot.FlowName = "profile"
	FlowRegistration         flowpilot.FlowName = "registration"
	FlowUserDetails          flowpilot.FlowName = "user_details"
	FlowMFAUsage             flowpilot.FlowName = "mfa_usage"
)
