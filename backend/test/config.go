package test

import "github.com/teamhanko/hanko/backend/config"

var DefaultConfig = config.Config{
	Webauthn: config.WebauthnSettings{
		RelyingParty: config.RelyingParty{
			Id:          "localhost",
			DisplayName: "Test Relying Party",
			Icon:        "",
			Origins:     []string{"http://localhost:8080"},
		},
		Timeout: 60000,
	},
	Secrets: config.Secrets{
		Keys: []string{"abcdefghijklmnop"},
	},
	Passcode: config.Passcode{Smtp: config.SMTP{
		Host: "localhost",
		Port: "2500",
	}},
	Session: config.Session{
		Lifespan: "1h",
		Cookie: config.Cookie{
			SameSite: "none",
		},
	},
	OIDC: config.OIDC{
		Enabled: false,
		Issuer:  "https://example.hanko.io",
		Key:     "gXK9jVVoRw6m85-XJHdSapaOPnBeifcJ6xcUxC-pJFk=",
		Clients: []config.OIDCClient{
			{
				ClientID:     "19286ac4-2216-44dd-bb21-02a41ea3548d",
				ClientSecret: "104cff48ae574505874884973de1f2488b8cd56ea55fdd45b2649a071af94617",
				ClientType:   "web",
				RedirectURI:  []string{"http://localhost:8080/callback"},
			},
		},
	},
}
