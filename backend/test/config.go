package test

import "github.com/teamhanko/hanko/backend/v2/config"

var DefaultConfig = config.Config{
	ApplicationConfig: config.ApplicationConfig{
		SecretKeys: []string{"abcdefghijklmnop"},
	},
	TenantConfig: config.TenantConfig{
		Webauthn: config.WebauthnSettings{
			RelyingParty: config.RelyingParty{
				Id:          "localhost",
				DisplayName: "Test Relying Party",
				Icon:        "",
				Origins:     []string{"http://localhost:8080", "http://localhost:8888"},
			},
			Timeouts: config.WebauthnTimeouts{
				Registration: 600000,
				Login:        600000,
			},
		},
		Secrets: config.Secrets{
			Keys: []string{"abcdefghijklmnop"},
			KeyManagement: config.KeyManagement{
				Type: "local",
			},
		},
		Email: config.Email{
			Enabled:              true,
			UseForAuthentication: true,
			PasscodeTtl:          300,
		},
		EmailDelivery: config.EmailDelivery{
			Enabled: true,
			SMTP: config.SMTP{
				Host: "localhost",
				Port: "2500",
			},
			FromAddress: "test@hanko.io",
			FromName:    "Hanko Test",
		},
		Session: config.Session{
			Lifespan: "1h",
			Cookie: config.Cookie{
				SameSite: "none",
			},
			Limit: 5,
		},
		Service: config.Service{
			Name: "Test",
		},
		Account: config.Account{
			AllowSignup:   true,
			AllowDeletion: false,
		},
		Passkey: config.Passkey{
			Enabled:          true,
			UserVerification: "preferred",
		},
	},
}
