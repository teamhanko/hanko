package login

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"reflect"
	"testing"
)

func TestContinueWithLoginIdentifier_generateFlow(t *testing.T) {
	type fields struct {
		Action shared.Action
	}
	type args struct {
		passkeysAcquireOnLogin string
		passwordAcquireOnLogin string
		hasPasskey             bool
		hasPassword            bool
		passkeyOptional        bool
		passwordOptional       bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []flowpilot.StateName
	}{
		// Test cases covering all possible combinations
		{
			name: "Case always-always (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "always",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding", "password_onboarding"},
		},
		{
			name: "Case always-always (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "always",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"password_onboarding"},
		},
		{
			name: "Case always-always (has password)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "always",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding"},
		},
		{
			name: "Case always-always (has both)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "always",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-conditional (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding"},
		},
		{
			name: "Case always-conditional (has password)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding"},
		},
		{
			name: "Case always-conditional (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-conditional (has both)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-always (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "always",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"password_onboarding"},
		},
		{
			name: "Case conditional-always (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "always",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"password_onboarding"},
		},
		{
			name: "Case conditional-always (has password)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "always",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-always (has both)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "always",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-conditional (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding", "password_onboarding"},
		},
		{
			name: "Case conditional-conditional (no credentials / both optional)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        true,
				passwordOptional:       true,
			},
			want: []flowpilot.StateName{"login_method_onboarding_chooser"},
		},
		{
			name: "Case conditional-conditional (no credentials / passkey required)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       true,
			},
			want: []flowpilot.StateName{"passkey_onboarding", "password_onboarding"},
		},
		{
			name: "Case conditional-conditional (no credentials / password required)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        true,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"password_onboarding", "passkey_onboarding"},
		},
		{
			name: "Case conditional-conditional (no credentials / both required)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding", "password_onboarding"},
		},
		{
			name: "Case conditional-conditional (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-conditional (has password)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-conditional (has both)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-never (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "never",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding"},
		},
		{
			name: "Case conditional-never (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "never",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-never (has password)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "never",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-never (has both)",
			args: args{
				passkeysAcquireOnLogin: "conditional",
				passwordAcquireOnLogin: "never",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-conditional (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"password_onboarding"},
		},
		{
			name: "Case never-conditional (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-conditional (has password)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-conditional (has both)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "conditional",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "never",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "never",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (has password)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "never",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (has both)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "never",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-always (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "always",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"password_onboarding"},
		},
		{
			name: "Case never-always (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "always",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"password_onboarding"},
		},
		{
			name: "Case never-always (has password)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "always",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-always (has both)",
			args: args{
				passkeysAcquireOnLogin: "never",
				passwordAcquireOnLogin: "always",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-never (no credentials)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "never",
				hasPasskey:             false,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding"},
		},
		{
			name: "Case always-never (has passkey)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "never",
				hasPasskey:             true,
				hasPassword:            false,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-never (has password)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "never",
				hasPasskey:             false,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{"passkey_onboarding"},
		},
		{
			name: "Case always-never (has both)",
			args: args{
				passkeysAcquireOnLogin: "always",
				passwordAcquireOnLogin: "never",
				hasPasskey:             true,
				hasPassword:            true,
				passkeyOptional:        false,
				passwordOptional:       false,
			},
			want: []flowpilot.StateName{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ContinueWithLoginIdentifier{
				Action: tt.fields.Action,
			}
			if got := a.generateFlow(tt.args.passkeysAcquireOnLogin, tt.args.passwordAcquireOnLogin, tt.args.hasPasskey, tt.args.hasPassword, tt.args.passkeyOptional, tt.args.passwordOptional); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateFlow() = %v, want %v", got, tt.want)
			}
		})
	}
}
