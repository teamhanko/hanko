package shared

import "github.com/teamhanko/hanko/backend/flowpilot"

const (
	StashPathEmail                                 flowpilot.JSONManagerPath = "email"
	StashPathEmailVerified                         flowpilot.JSONManagerPath = "email_verified"
	StashPathLoginMethod                           flowpilot.JSONManagerPath = "login_method"
	StashPathNewPassword                           flowpilot.JSONManagerPath = "new_password"
	StashPathPasscodeEmail                         flowpilot.JSONManagerPath = "passcode_email"
	StashPathPasscodeID                            flowpilot.JSONManagerPath = "passcode_id"
	StashPathPasscodeTemplate                      flowpilot.JSONManagerPath = "passcode_template"
	StashPathSkipUserCreation                      flowpilot.JSONManagerPath = "skip_user_creation"
	StashPathUserHasPassword                       flowpilot.JSONManagerPath = "user_has_password"
	StashPathUserHasWebauthnCredential             flowpilot.JSONManagerPath = "user_has_webauthn_credential"
	StashPathUserID                                flowpilot.JSONManagerPath = "user_id"
	StashPathUsername                              flowpilot.JSONManagerPath = "username"
	StashPathWebauthnAvailable                     flowpilot.JSONManagerPath = "webauthn_available"
	StashPathWebauthnConditionalMediationAvailable flowpilot.JSONManagerPath = "webauthn_conditional_mediation_available"
	StashPathWebauthnCredential                    flowpilot.JSONManagerPath = "webauthn_credential"
	StashPathWebauthnSessionDataID                 flowpilot.JSONManagerPath = "webauthn_session_data_id"
)
