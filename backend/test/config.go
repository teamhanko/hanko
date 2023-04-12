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
}
