package shared

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	FlowCapabilities         flowpilot.FlowName = "capabilities"
	FlowCredentialOnboarding flowpilot.FlowName = "credential_onboarding"
	FlowCredentialUsage      flowpilot.FlowName = "credential_usage"
	FlowLogin                flowpilot.FlowName = "login"
	FlowMFACreation          flowpilot.FlowName = "mfa_creation"
	FlowProfile              flowpilot.FlowName = "profile"
	FlowRegistration         flowpilot.FlowName = "registration"
	FlowUserDetails          flowpilot.FlowName = "user_details"
	FlowMFAUsage             flowpilot.FlowName = "mfa_usage"
)
