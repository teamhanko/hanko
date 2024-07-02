package login

import (
	"github.com/teamhanko/hanko/backend/config"
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
		cfg         config.Config
		hasPasskey  bool
		hasPassword bool
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
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey", "password_creation"},
		},
		{
			name: "Case always-always (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"password_creation"},
		},
		{
			name: "Case always-always (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey"},
		},
		{
			name: "Case always-always (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-conditional (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey"},
		},
		{
			name: "Case always-conditional (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey"},
		},
		{
			name: "Case always-conditional (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-conditional (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-always (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"password_creation"},
		},
		{
			name: "Case conditional-always (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"password_creation"},
		},
		{
			name: "Case conditional-always (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-always (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-conditional (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey", "password_creation"},
		},
		{
			name: "Case conditional-conditional (no credentials / both optional)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: true},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: true},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"login_method_onboarding_chooser"},
		},
		{
			name: "Case conditional-conditional (no credentials / passkey required)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: true},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey", "password_creation"},
		},
		{
			name: "Case conditional-conditional (no credentials / password required)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: true},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"password_creation", "onboarding_create_passkey"},
		},
		{
			name: "Case conditional-conditional (no credentials / both required)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey", "password_creation"},
		},
		{
			name: "Case conditional-conditional (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-conditional (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-conditional (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-never (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey"},
		},
		{
			name: "Case conditional-never (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-never (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case conditional-never (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "conditional", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-conditional (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"password_creation"},
		},
		{
			name: "Case never-conditional (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-conditional (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-conditional (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "conditional", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-never (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-always (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"password_creation"},
		},
		{
			name: "Case never-always (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"password_creation"},
		},
		{
			name: "Case never-always (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case never-always (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "never", Optional: false},
					Password: config.Password{AcquireOnLogin: "always", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-never (no credentials)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: false,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey"},
		},
		{
			name: "Case always-never (has passkey)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: false,
			},
			want: []flowpilot.StateName{},
		},
		{
			name: "Case always-never (has password)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  false,
				hasPassword: true,
			},
			want: []flowpilot.StateName{"onboarding_create_passkey"},
		},
		{
			name: "Case always-never (has both)",
			args: args{
				cfg: config.Config{
					Passkey:  config.Passkey{AcquireOnLogin: "always", Optional: false},
					Password: config.Password{AcquireOnLogin: "never", Optional: false},
				},
				hasPasskey:  true,
				hasPassword: true,
			},
			want: []flowpilot.StateName{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ContinueWithLoginIdentifier{
				Action: tt.fields.Action,
			}
			if got := a.determineCredentialOnboardingStates(tt.args.cfg, tt.args.hasPasskey, tt.args.hasPassword); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("determineCredentialOnboardingStates() = %v, want %v", got, tt.want)
			}
		})
	}
}
